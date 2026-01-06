#!/bin/bash
# Quick check if Ralph is making progress

echo "╔═══════════════════════════════════════╗"
echo "║   Ralph Progress Check                ║"
echo "╚═══════════════════════════════════════╝"
echo ""

echo "📝 Files changed since last commit:"
git status --short || echo "  (no changes)"
echo ""

echo "📋 Tasks completed:"
completed=$(grep -c "^\- \[x\]" @fix_plan.md 2>/dev/null || echo "0")
total=$(grep -c "^\- \[.\]" @fix_plan.md 2>/dev/null || echo "85")
echo "  $completed / $total tasks"
echo ""

echo "🎯 Feature status:"
echo -n "  T1 Goodbye packets:     "
grep -q "BuildGoodbyeRecords" internal/records/record_set.go 2>/dev/null && echo "✅" || echo "❌"

echo -n "  T2 Source validation:   "
grep -q "validateSourceAddress" responder/responder.go 2>/dev/null && echo "✅" || echo "❌"

echo -n "  T3 TC bit:              "
grep -q "0x02.*TC" internal/responder/response_builder.go 2>/dev/null && echo "✅" || echo "❌"

echo -n "  T4 QU bit:              "
grep -q "QU.*bit" responder/responder.go 2>/dev/null && echo "✅" || echo "❌"
echo ""

echo "🚦 Exit signals:"
if [ -f .exit_signals ]; then
    done_count=$(jq '.done_signals | length' .exit_signals 2>/dev/null || echo "0")
    echo "  $done_count consecutive 'done' signals (need 10 to exit)"
else
    echo "  (no file)"
fi
echo ""

if git status --short | grep -q "\.go$"; then
    echo "✅ Ralph is writing code!"
else
    echo "⚠️  No .go files changed - Ralph might be stuck!"
fi
