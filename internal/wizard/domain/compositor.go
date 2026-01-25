package domain

// Compositor represents supported Wayland compositors.
type Compositor string

const (
	CompositorNiri     Compositor = "niri"
	CompositorHyprland Compositor = "hyprland"
)
