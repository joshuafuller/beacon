// Package message implements DNS name encoding and compression per RFC 1035 §4.1.4.
package message

import (
	"fmt"
	"strings"

	"github.com/joshuafuller/beacon/internal/errors"
	"github.com/joshuafuller/beacon/internal/protocol"
)

// ParseName parses a DNS name from a message buffer, handling compression pointers
// per RFC 1035 §4.1.4.
//
// DNS names are encoded as a sequence of labels. Each label is prefixed by a length byte.
// A zero-length label (0x00) terminates the name.
//
// RFC 1035 §4.1.4 defines message compression: labels can be replaced by a pointer
// to a prior occurrence of the same name. A pointer is indicated by the two high-order
// bits being set (0xC0), followed by a 14-bit offset.
//
// This function detects compression loops by limiting the number of pointer jumps
// to MaxCompressionPointers (256).
//
// FR-012: System MUST decompress DNS names per RFC 1035 §4.1.4 (message compression)
//
// Parameters:
//   - msg: The complete DNS message buffer (needed for following compression pointers)
//   - offset: The starting offset of the name in the buffer
//
// Returns:
//   - name: The decompressed DNS name (e.g., "printer.local")
//   - newOffset: The offset immediately after the name (for parsing subsequent fields)
//   - error: WireFormatError if the name is malformed
func ParseName(msg []byte, offset int) (name string, newOffset int, err error) {
	if offset < 0 || offset >= len(msg) {
		return "", offset, &errors.WireFormatError{
			Operation: "parse name",
			Offset:    offset,
			Message:   "offset out of bounds",
		}
	}

	var labels []string
	jumps := 0
	pos := offset
	jumped := false

	for {
		// Check bounds
		if pos >= len(msg) {
			return "", offset, &errors.WireFormatError{
				Operation: "parse name",
				Offset:    pos,
				Message:   "unexpected end of message while parsing name",
			}
		}

		length := msg[pos]

		// Check for compression pointer per RFC 1035 §4.1.4
		if (length & protocol.CompressionMask) == protocol.CompressionMask {
			// Compression pointer (high 2 bits = 11)
			if pos+1 >= len(msg) {
				return "", offset, &errors.WireFormatError{
					Operation: "parse name",
					Offset:    pos,
					Message:   "truncated compression pointer",
				}
			}

			// Extract 14-bit offset: combine two bytes and mask out high 2 bits
			pointerOffset := int(msg[pos]&0x3F)<<8 | int(msg[pos+1])

			// Validate pointer doesn't point forward (RFC 1035 §4.1.4: pointers point backwards)
			if pointerOffset >= pos {
				return "", offset, &errors.WireFormatError{
					Operation: "parse name",
					Offset:    pos,
					Message:   fmt.Sprintf("invalid compression pointer: points to offset %d (current position %d)", pointerOffset, pos),
				}
			}

			// Update newOffset only on first jump (subsequent jumps don't affect wire position)
			if !jumped {
				newOffset = pos + 2
				jumped = true
			}

			// Follow the pointer
			pos = pointerOffset

			// Detect compression loops per FR-012
			jumps++
			if jumps > protocol.MaxCompressionPointers {
				return "", offset, &errors.WireFormatError{
					Operation: "parse name",
					Offset:    pos,
					Message:   fmt.Sprintf("too many compression jumps (possible loop, exceeded %d jumps)", protocol.MaxCompressionPointers),
				}
			}

			continue
		}

		// Check for terminator (zero-length label)
		if length == 0 {
			// End of name
			if !jumped {
				newOffset = pos + 1
			}
			break
		}

		// Validate label length per RFC 1035 §3.1
		if length > protocol.MaxLabelLength {
			return "", offset, &errors.WireFormatError{
				Operation: "parse name",
				Offset:    pos,
				Message:   fmt.Sprintf("label length %d exceeds maximum %d bytes per RFC 1035 §3.1", length, protocol.MaxLabelLength),
			}
		}

		// Check if we have enough bytes for this label
		if pos+1+int(length) > len(msg) {
			return "", offset, &errors.WireFormatError{
				Operation: "parse name",
				Offset:    pos,
				Message:   fmt.Sprintf("truncated label: expected %d bytes, only %d available", length, len(msg)-pos-1),
			}
		}

		// Extract label
		label := string(msg[pos+1 : pos+1+int(length)])
		labels = append(labels, label)

		// Move to next label
		pos += 1 + int(length)
	}

	// Join labels with dots to form the complete name
	name = strings.Join(labels, ".")

	// Validate total name length per RFC 1035 §3.1
	// Note: Wire format length includes length bytes, but MaxNameLength applies to the string representation
	if len(name) > protocol.MaxNameLength {
		return "", offset, &errors.WireFormatError{
			Operation: "parse name",
			Offset:    offset,
			Message:   fmt.Sprintf("name length %d exceeds maximum %d bytes per RFC 1035 §3.1", len(name), protocol.MaxNameLength),
		}
	}

	return name, newOffset, nil
}

// EncodeServiceInstanceName encodes a DNS name into wire format per RFC 1035 §3.1.
//
// The name is split into labels (separated by dots), and each label is prefixed
// by its length byte. A zero-length label (0x00) terminates the name.
//
// RFC 1035 §3.1: Labels are sequences of ASCII characters, length-prefixed.
// Example: "printer.local" → [7]printer[5]local[0]
//
// M1 does NOT implement compression (compression is SHOULD, not MUST per RFC 6762 §18.14).
// Compression is deferred to future milestones for simplicity.
//
// FR-003: System MUST validate queried names follow DNS naming rules (labels ≤63 bytes, total name ≤255 bytes)
//
// Parameters:
//   - name: The DNS name to encode (e.g., "printer.local")
//
// Returns:
//   - encoded: The wire format representation
//   - error: ValidationError if the name is invalid
//
// Complexity 21 due to RFC 1035 §3.1 DNS name encoding requirements (label parsing, character validation, compression handling, length constraints)
// EncodeServiceInstanceName encodes a service instance name per RFC 6763 §4.3.
//
// RFC 6763 §4.3: Service instance names use length-prefixed labels where the instance
// portion is a SINGLE label that can contain arbitrary UTF-8 characters including spaces.
//
// Example: "My Printer._http._tcp.local" is encoded as:
//
//	[10]My Printer[5]_http[4]_tcp[5]local[0]
//
// Parameters:
//   - instanceName: User-friendly instance name (can contain spaces, UTF-8)
//   - serviceType: Service type (e.g., "_http._tcp.local")
//
// Returns: Fully encoded DNS name (instance.servicetype)
func EncodeServiceInstanceName(instanceName, serviceType string) ([]byte, error) {
	if len(instanceName) == 0 {
		return nil, &errors.ValidationError{
			Field:   "instanceName",
			Value:   instanceName,
			Message: "instance name cannot be empty",
		}
	}

	if len(instanceName) > protocol.MaxLabelLength {
		return nil, &errors.ValidationError{
			Field:   "instanceName",
			Value:   instanceName,
			Message: fmt.Sprintf("instance name exceeds maximum label length %d bytes", protocol.MaxLabelLength),
		}
	}

	// Encode instance name as a single label (allow spaces and UTF-8)
	encoded := make([]byte, 0, 256)
	encoded = append(encoded, byte(len(instanceName))) // Length prefix
	encoded = append(encoded, []byte(instanceName)...) // Raw bytes (UTF-8)

	// Encode service type normally (strict DNS validation)
	serviceTypeEncoded, err := EncodeName(serviceType)
	if err != nil {
		return nil, fmt.Errorf("encoding service type: %w", err)
	}

	// Concatenate: instance label + service type labels
	// Remove trailing null from serviceTypeEncoded (we'll add it at the end)
	if len(serviceTypeEncoded) > 0 && serviceTypeEncoded[len(serviceTypeEncoded)-1] == 0 {
		serviceTypeEncoded = serviceTypeEncoded[:len(serviceTypeEncoded)-1]
	}

	encoded = append(encoded, serviceTypeEncoded...)
	encoded = append(encoded, 0) // Null terminator

	return encoded, nil
}

// EncodeName encodes a DNS name per RFC 1035 §3.1.
func EncodeName(name string) ([]byte, error) {
	// Handle empty name (root ".")
	if name == "" || name == "." {
		return []byte{0}, nil
	}

	// Split into labels
	labels := strings.Split(name, ".")

	// Remove trailing empty label if name ends with "."
	if len(labels) > 0 && labels[len(labels)-1] == "" {
		labels = labels[:len(labels)-1]
	}

	// Validate and encode
	encoded := make([]byte, 0, 256) // Pre-allocate typical DNS name size (max 255 bytes)
	for _, label := range labels {
		if err := validateLabel(label, name); err != nil {
			return nil, err
		}
		encoded = append(encoded, byte(len(label)))
		encoded = append(encoded, label...)
	}

	// Append terminator (zero-length label)
	encoded = append(encoded, 0)

	// Validate total encoded length
	// Note: MaxNameLength (255) applies to the total wire format, including length bytes
	if len(encoded) > protocol.MaxNameLength {
		return nil, &errors.ValidationError{
			Field:   "name",
			Value:   name,
			Message: fmt.Sprintf("encoded name length %d exceeds maximum %d bytes per RFC 1035 §3.1", len(encoded), protocol.MaxNameLength),
		}
	}

	return encoded, nil
}

// validateLabel checks a single DNS label per RFC 1035 §3.1.
func validateLabel(label, name string) error {
	if len(label) == 0 {
		return &errors.ValidationError{Field: "name", Value: name, Message: "empty label (consecutive dots)"}
	}
	if len(label) > protocol.MaxLabelLength {
		return &errors.ValidationError{Field: "name", Value: name, Message: fmt.Sprintf("label %q exceeds maximum length %d bytes per RFC 1035 §3.1", label, protocol.MaxLabelLength)}
	}
	for i, ch := range label {
		valid := (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '-' || ch == '_'
		if !valid {
			return &errors.ValidationError{Field: "name", Value: name, Message: fmt.Sprintf("invalid character %q in label %q (position %d)", ch, label, i)}
		}
		if ch == '-' && (i == 0 || i == len(label)-1) {
			return &errors.ValidationError{Field: "name", Value: name, Message: fmt.Sprintf("hyphen cannot be first or last character in label %q", label)}
		}
	}
	return nil
}
