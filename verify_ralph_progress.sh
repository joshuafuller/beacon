#!/bin/bash
# Quick script to verify Ralph is actually making progress

echo "=== Ralph Progress Verification ==="
echo ""

echo "1. Files changed since last commit:"
git status --short
echo ""

echo "2. Completed tasks in @fix_plan.md:"
completed=$(grep -c "^\- \[x\]" @fix_plan.md 2>/dev/null || echo "0")
total=$(grep -c "^\- \[" @fix_plan.md 2>/dev/null || echo "85")
echo "   $completed / $total tasks complete"
echo ""

echo "3. Feature implementation status:"
echo -n "   Goodbye packets: "
grep -q "BuildGoodbyeRecords" internal/records/record_set.go 2>/dev/null && echo "✅ FOUND" || echo "❌ MISSING"

echo -n "   Source validation: "
grep -q "validateSourceAddress" responder/responder.go 2>/dev/null && echo "✅ FOUND" || echo "❌ MISSING"

echo -n "   TC bit: "
grep -q "TC.*bit.*0x02" internal/responder/response_builder.go 2>/dev/null && echo "✅ FOUND" || echo "❌ MISSING"

echo -n "   QU bit: "
grep -q "QU.*bit" responder/responder.go 2>/dev/null && echo "✅ FOUND" || echo "❌ MISSING"
echo ""

echo "4. Exit signals:"
cat .exit_signals 2>/dev/null || echo "   (no file)"
echo ""

echo "=== If 'Files changed' is empty, Ralph is NOT working! ==="
