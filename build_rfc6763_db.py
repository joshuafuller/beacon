#!/usr/bin/env python3
"""
Build COMPLETE RFC 6763 requirements database (all sections 1-16).
"""

import re
import json
import os

def parse_rfc_sections(rfc_path):
    """Parse RFC 6763 into structured sections."""
    with open(rfc_path, 'r', encoding='utf-8', errors='replace') as f:
        lines = f.readlines()

    # Remove page headers/footers
    cleaned = []
    for line in lines:
        if 'Cheshire & Krochmal' in line or \
           (line.strip().startswith('RFC 6763') and 'DNS-Based Service Discovery' in line) or \
           line.strip().startswith('[Page '):
            continue
        cleaned.append(line)

    # Find section boundaries
    section_pattern = re.compile(r'^(\d+(?:\.\d+)*)\.\s+(.+)$')
    sections = []
    current_section = None
    current_lines = []

    for i, line in enumerate(cleaned):
        match = section_pattern.match(line.strip())
        if match:
            # Save previous section
            if current_section:
                sections.append({
                    'number': current_section[0],
                    'title': current_section[1],
                    'text': ''.join(current_lines),
                    'line_start': current_section[2]
                })
            current_section = (match.group(1), match.group(2), i)
            current_lines = []
        elif current_section:
            current_lines.append(line)

    # Add last section
    if current_section:
        sections.append({
            'number': current_section[0],
            'title': current_section[1],
            'text': ''.join(current_lines),
            'line_start': current_section[2]
        })

    return sections

def extract_normative_statements(text, section_num, section_title):
    """Extract all RFC 2119 normative statements from text."""
    requirements = []

    # RFC 2119 keywords patterns - more precise
    patterns = {
        'MUST NOT': r'\bMUST NOT\b|\bSHALL NOT\b',
        'MUST': r'\bMUST\b|\bSHALL\b|\bREQUIRED\b',
        'SHOULD NOT': r'\bSHOULD NOT\b',
        'SHOULD': r'\bSHOULD\b|\bRECOMMENDED\b',
        'MAY': r'\bMAY\b|\bOPTIONAL\b',
    }

    # Split into sentences
    text = text.replace('e.g.,', 'e_g_')
    text = text.replace('i.e.,', 'i_e_')
    text = text.replace('etc.', 'etc_')
    text = text.replace('vs.', 'vs_')

    # Split on sentence boundaries
    sentences = re.split(r'(?<=[.!?])\s+(?=[A-Z"\(])', text)

    for sent in sentences:
        # Restore abbreviations
        sent = sent.replace('e_g_', 'e.g.,')
        sent = sent.replace('i_e_', 'i.e.,')
        sent = sent.replace('etc_', 'etc.')
        sent = sent.replace('vs_', 'vs.')

        # Check for normative keywords (in priority order)
        req_type = None
        for rtype, pattern in patterns.items():
            if re.search(pattern, sent):
                req_type = rtype
                break

        if req_type:
            # Clean up sentence
            sent = ' '.join(sent.split())
            # Skip very short or header-like matches
            if len(sent) > 30 and not sent.startswith('Table of Contents'):
                requirements.append({
                    'type': req_type,
                    'text': sent,
                    'section': section_num,
                    'section_title': section_title
                })

    return requirements

def search_beacon_implementation(req_text, beacon_root):
    """Search Beacon codebase for potential implementation."""

    # Map requirement keywords to search terms (DNS-SD specific)
    keyword_map = {
        'service instance': ['Service', 'service', 'Instance', 'InstanceName'],
        'service type': ['ServiceType', 'serviceType', '_tcp', '_udp'],
        'PTR': ['PTR', 'ptrRecord'],
        'SRV': ['SRV', 'srvRecord'],
        'TXT': ['TXT', 'txtRecord', 'TXTRecord'],
        'txt record': ['TXT', 'txtRecord', 'TXTRecord'],
        'key/value': ['KeyValue', 'key-value', 'TXT'],
        'instance name': ['InstanceName', 'instanceName'],
        'service name': ['ServiceType', 'serviceType'],
        'domain': ['Domain', 'domain'],
        'browse': ['Browse', 'browsing', 'enumeration'],
        'enumeration': ['Enumerate', 'enumeration', 'Browse'],
        'subtype': ['Subtype', 'subtype'],
        'unicast': ['unicast', 'Unicast'],
        'multicast': ['multicast', 'Multicast'],
        'query': ['Query', 'query'],
        'response': ['Response', 'response'],
        'additional': ['Additional', 'additional'],
        'record': ['Record', 'RR'],
        'label': ['Label', 'label', 'DNS name'],
        'encoding': ['Encode', 'encoding'],
        'utf-8': ['UTF8', 'utf8', 'Unicode'],
        'compression': ['compression', 'compress'],
        'port': ['Port', 'port'],
        'target': ['Target', 'target'],
        'priority': ['Priority', 'priority'],
        'weight': ['Weight', 'weight'],
        'ttl': ['TTL', 'ttl', 'TimeToLive'],
        'cache': ['cache', 'Cache'],
    }

    search_terms = set()
    text_lower = req_text.lower()

    for keyword, terms in keyword_map.items():
        if keyword in text_lower:
            search_terms.update(terms)

    if not search_terms:
        # Fallback: extract capitalized words
        words = re.findall(r'\b[A-Z][a-z]+\b', req_text)
        search_terms.update(words[:3])  # Top 3

    implementations = set()
    tests = set()

    # Key source directories
    source_dirs = [
        'querier',
        'responder',
        'internal/transport',
        'internal/responder',
        'internal/state',
        'internal/records',
        'internal/protocol',
        'internal/message',
        'internal/security'
    ]

    for dir_name in source_dirs:
        dir_path = os.path.join(beacon_root, dir_name)
        if not os.path.exists(dir_path):
            continue

        for root, dirs, files in os.walk(dir_path):
            for file in files:
                if not file.endswith('.go'):
                    continue

                file_path = os.path.join(root, file)
                try:
                    with open(file_path, 'r') as f:
                        content = f.read()

                        for term in search_terms:
                            if term in content:
                                rel_path = os.path.relpath(file_path, beacon_root)
                                if file.endswith('_test.go'):
                                    tests.add(rel_path)
                                else:
                                    implementations.add(rel_path)
                                break
                except:
                    pass

    return sorted(list(implementations)), sorted(list(tests))

def main():
    beacon_root = '/home/user/development/beacon'
    rfc_path = os.path.join(beacon_root, 'RFC Docs/RFC-6763-DNS-SD.txt')

    print("Parsing RFC 6763...")
    sections = parse_rfc_sections(rfc_path)
    print(f"Found {len(sections)} sections")

    # Extract requirements from sections 1-16 (normative sections)
    all_requirements = []
    req_id = 1

    for section in sections:
        sec_num = section['number']
        # Only process main sections 1-16 (normative)
        if sec_num and sec_num.split('.')[0].isdigit():
            main_section = int(sec_num.split('.')[0])
            if 1 <= main_section <= 16:
                requirements = extract_normative_statements(
                    section['text'],
                    sec_num,
                    section['title']
                )

                for req in requirements:
                    req['id'] = req_id
                    req['rfc_id'] = f"RFC6763-§{sec_num}-REQ-{req_id:03d}"
                    req_id += 1
                    all_requirements.append(req)

    print(f"Extracted {len(all_requirements)} requirements from sections 1-16")

    # Search for implementations
    print("Searching for implementations...")
    for i, req in enumerate(all_requirements):
        if i % 20 == 0:
            print(f"  Processing requirement {i+1}/{len(all_requirements)}...")

        implementations, tests = search_beacon_implementation(req['text'], beacon_root)
        req['implementations'] = implementations
        req['tests'] = tests

        # Determine status
        if implementations and tests:
            req['status'] = 'COMPLETE'
        elif implementations:
            req['status'] = 'PARTIAL'
        else:
            req['status'] = 'MISSING'

        # Set priority
        if 'MUST' in req['type']:
            req['priority'] = 'P0'
        elif 'SHOULD' in req['type']:
            req['priority'] = 'P1'
        else:
            req['priority'] = 'P2'

    # Generate markdown report
    print("Generating comprehensive report...")

    with open(os.path.join(beacon_root, 'RFC6763_REQUIREMENTS_COMPLETE.md'), 'w') as f:
        f.write("# RFC 6763 Complete Requirements Database\n\n")
        f.write(f"**Generated**: 2026-01-06\n\n")
        f.write("**Scope**: All sections 1-16 of RFC 6763 (DNS-Based Service Discovery)\n\n")

        # Summary statistics
        must_count = len([r for r in all_requirements if r['type'] == 'MUST'])
        must_not_count = len([r for r in all_requirements if r['type'] == 'MUST NOT'])
        should_count = len([r for r in all_requirements if r['type'] == 'SHOULD'])
        should_not_count = len([r for r in all_requirements if r['type'] == 'SHOULD NOT'])
        may_count = len([r for r in all_requirements if r['type'] == 'MAY'])

        complete_count = len([r for r in all_requirements if r['status'] == 'COMPLETE'])
        partial_count = len([r for r in all_requirements if r['status'] == 'PARTIAL'])
        missing_count = len([r for r in all_requirements if r['status'] == 'MISSING'])

        f.write("## Summary\n\n")
        f.write(f"**Total Requirements**: {len(all_requirements)}\n\n")
        f.write("### By Type\n")
        f.write(f"- **MUST**: {must_count} (P0 - Mandatory)\n")
        f.write(f"- **MUST NOT**: {must_not_count} (P0 - Prohibited)\n")
        f.write(f"- **SHOULD**: {should_count} (P1 - Strong Recommendation)\n")
        f.write(f"- **SHOULD NOT**: {should_not_count} (P1 - Not Recommended)\n")
        f.write(f"- **MAY**: {may_count} (P2 - Optional)\n\n")

        f.write("### Implementation Status\n")
        f.write(f"- ✅ **Complete**: {complete_count} ({complete_count*100//len(all_requirements)}%)\n")
        f.write(f"- ⚠️  **Partial**: {partial_count} ({partial_count*100//len(all_requirements)}%)\n")
        f.write(f"- ❌ **Missing**: {missing_count} ({missing_count*100//len(all_requirements)}%)\n\n")

        # P0 (MUST) Gap Analysis
        p0_missing = [r for r in all_requirements if r['priority'] == 'P0' and r['status'] == 'MISSING']
        p0_partial = [r for r in all_requirements if r['priority'] == 'P0' and r['status'] == 'PARTIAL']

        f.write("### P0 (MUST) Gap Analysis\n\n")
        f.write(f"- Total P0 requirements: {must_count + must_not_count}\n")
        f.write(f"- ❌ Missing: {len(p0_missing)}\n")
        f.write(f"- ⚠️  Partial: {len(p0_partial)}\n")
        f.write(f"- ✅ Complete: {must_count + must_not_count - len(p0_missing) - len(p0_partial)}\n\n")

        if p0_missing:
            f.write("**Critical Missing P0 Requirements**:\n")
            for req in p0_missing[:10]:  # Top 10
                f.write(f"- {req['rfc_id']}: {req['text'][:100]}...\n")
            f.write("\n")

        # Group by section
        sections_dict = {}
        for req in all_requirements:
            sec = req['section']
            if sec not in sections_dict:
                sections_dict[sec] = []
            sections_dict[sec].append(req)

        f.write("---\n\n")
        f.write("## Requirements by Section\n\n")

        for sec in sorted(sections_dict.keys(), key=lambda x: [int(n) if n.isdigit() else 0 for n in x.split('.')]):
            reqs = sections_dict[sec]
            section_title = reqs[0]['section_title']

            # Section summary
            sec_complete = len([r for r in reqs if r['status'] == 'COMPLETE'])
            sec_total = len(reqs)

            f.write(f"### §{sec} {section_title}\n\n")
            f.write(f"**Progress**: {sec_complete}/{sec_total} complete ({sec_complete*100//sec_total if sec_total > 0 else 0}%)\n\n")

            for req in reqs:
                status_icon = {'COMPLETE': '✅', 'PARTIAL': '⚠️', 'MISSING': '❌'}[req['status']]
                f.write(f"#### {req['rfc_id']} {status_icon}\n\n")
                f.write(f"- **Type**: {req['type']}\n")
                f.write(f"- **Priority**: {req['priority']}\n")
                f.write(f"- **Status**: {req['status']}\n\n")
                f.write(f"**Requirement**:\n")
                f.write(f"> {req['text']}\n\n")

                if req['implementations']:
                    f.write(f"**Implementation**:\n")
                    for impl in req['implementations']:
                        f.write(f"- `{impl}`\n")
                    f.write("\n")
                else:
                    f.write(f"**Implementation**: NOT IMPLEMENTED\n\n")

                if req['tests']:
                    f.write(f"**Tests**:\n")
                    for test in req['tests']:
                        f.write(f"- `{test}`\n")
                    f.write("\n")
                else:
                    f.write(f"**Tests**: NO TEST\n\n")

                f.write("---\n\n")

    # Save JSON
    with open(os.path.join(beacon_root, 'rfc6763_requirements_complete.json'), 'w') as f:
        json.dump(all_requirements, f, indent=2)

    print(f"✅ Generated RFC6763_REQUIREMENTS_COMPLETE.md")
    print(f"✅ Generated rfc6763_requirements_complete.json")
    print(f"\n📊 Final Summary:")
    print(f"  Total Requirements: {len(all_requirements)}")
    print(f"  MUST: {must_count}, MUST NOT: {must_not_count}")
    print(f"  SHOULD: {should_count}, SHOULD NOT: {should_not_count}")
    print(f"  MAY: {may_count}")
    print(f"  ✅ Complete: {complete_count} ({complete_count*100//len(all_requirements)}%)")
    print(f"  ⚠️  Partial: {partial_count} ({partial_count*100//len(all_requirements)}%)")
    print(f"  ❌ Missing: {missing_count} ({missing_count*100//len(all_requirements)}%)")
    print(f"\n🚨 P0 Gaps: {len(p0_missing)} missing, {len(p0_partial)} partial")

if __name__ == '__main__':
    main()
