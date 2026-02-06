#!/bin/bash
# Generate RFC Compliance Badge for README
# Uses shields.io dynamic badge generation

set -euo pipefail

BEACON_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REQUIREMENTS_FILE="${BEACON_ROOT}/RFC_REQUIREMENTS_COMPLETE.md"

# Extract compliance percentage
compliance_pct=$(grep "^- ✅ \*\*Complete\*\*: " "$REQUIREMENTS_FILE" | head -1 | sed 's/.*(\([0-9]*\)%).*/\1/')

# Extract P0 gaps
p0_gaps=$(grep "^- ❌ Missing: " "$REQUIREMENTS_FILE" | head -1 | awk '{print $4}')

# Determine badge color
if [[ $compliance_pct -ge 95 ]]; then
    color="brightgreen"
elif [[ $compliance_pct -ge 85 ]]; then
    color="green"
elif [[ $compliance_pct -ge 75 ]]; then
    color="yellowgreen"
else
    color="orange"
fi

# P0 badge color
if [[ $p0_gaps -eq 0 ]]; then
    p0_color="brightgreen"
    p0_text="100%25%20P0"
else
    p0_color="red"
    p0_text="${p0_gaps}%20P0%20gaps"
fi

# Generate badge URLs
overall_badge="![RFC Compliance](https://img.shields.io/badge/RFC%206762-${compliance_pct}%25-${color}?style=flat-square&logo=checkmarx)"
p0_badge="![P0 Compliance](https://img.shields.io/badge/P0%20Compliance-${p0_text}-${p0_color}?style=flat-square&logo=checkmarx)"

# Output for README
echo "## RFC Compliance Badges"
echo ""
echo "Add these to your README.md:"
echo ""
echo "\`\`\`markdown"
echo "$overall_badge"
echo "$p0_badge"
echo "\`\`\`"
echo ""
echo "Or copy/paste these URLs:"
echo ""
echo "Overall: https://img.shields.io/badge/RFC%206762-${compliance_pct}%25-${color}?style=flat-square&logo=checkmarx"
echo "P0: https://img.shields.io/badge/P0%20Compliance-${p0_text}-${p0_color}?style=flat-square&logo=checkmarx"
