package main

// iconData returns the menu bar icon as PNG bytes
// Using a simple port/network icon (16x16 PNG, base64 encoded)
func iconData() []byte {
	// This is a placeholder - replace with actual icon
	// Simple 16x16 black dot for now
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		// ... (minimal PNG data for a simple icon)
	}
}
