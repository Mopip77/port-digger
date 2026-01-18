# Port Digger - Manual Test Checklist

## Edge Cases to Test

### Scanner
- [ ] No ports listening (stop all servers)
- [ ] Many ports (>20)
- [ ] Port number edge cases (80, 8080, 65535)
- [ ] Process names with spaces
- [ ] Multiple processes on different ports

### Actions
- [ ] Open browser on non-serving port (should fail gracefully)
- [ ] Copy port to clipboard, paste elsewhere
- [ ] Kill own process (should prompt for password)
- [ ] Kill process without sudo (should work for own processes)
- [ ] Kill process with sudo (test with system process)

### Menu
- [ ] Click refresh multiple times rapidly
- [ ] Open submenu, then refresh menu
- [ ] Hover over multiple ports quickly

### System
- [ ] Run on macOS 13+
- [ ] Check memory usage with Activity Monitor
- [ ] Leave running for extended period
- [ ] Check binary size

## Success Criteria
- No crashes
- Memory stays under 30MB
- UI remains responsive
- All actions work or fail gracefully
