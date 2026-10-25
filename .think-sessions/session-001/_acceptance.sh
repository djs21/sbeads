#!/bin/bash
set -e

echo "=== Testing sbeads binary ==="

# Test 1: Create an issue
echo "Test 1: Create an issue"
./main create "Test issue" --priority 1 > /tmp/sbeads_test.log 2>&1 || true
echo "Created issue log:"
cat /tmp/sbeads_test.log

# Test 2: Show issues
echo ""
echo "Test 2: Show issues"
./main show bd-a1b2 --issue-id bd-a1b2 > /tmp/show.log 2>&1 || true
echo "Show log:"
cat /tmp/show.log

# Test 3: List ready issues
echo ""
echo "Test 3: List ready issues"
./main ready > /tmp/ready.log 2>&1 || true
echo "Ready log:"
cat /tmp/ready.log

# Test 4: Close issue
echo ""
echo "Test 4: Close issue"
./main close bd-a1b2 "Test issue closed" --issue-id bd-a1b2 > /tmp/close.log 2>&1 || true
echo "Close log:"
cat /tmp/close.log

# Test 5: Dependencies
echo ""
echo "Test 5: Add dependency"
./main dep add bd-a1b2 bd-a2b3 --child bd-a1b2 --parent bd-a2b3 > /tmp/dep.log 2>&1 || true
echo "Add dependency log:"
cat /tmp/dep.log

echo ""
echo "=== All tests completed ==="

# Cleanup
rm -f /tmp/sbeads_test.log /tmp/show.log /tmp/ready.log /tmp/close.log /tmp/dep.log
