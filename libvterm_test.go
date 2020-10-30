package libvterm

import "testing"
import "fmt"

func TestInitialize(t *testing.T) {
	vt := New(100, 100)
	defer vt.Close()

	screen := vt.ObtainScreen()

	if err := screen.SetSbPushLineCallback(func(cols int, cells CellP) error {
		return nil
	}); err != nil {
		t.Error(err)
	}

	screen.Reset()
	if err := screen.Reset(); err != nil {
		t.Error(err)
	}

	fmt.Fprintln(vt, "Hello, World!")

	text := screen.GetText(Rect{0, 1, 0, 20})
	if text != "Hello, World!" {
		t.Errorf("text: \"%s\" is not equal to \"Hello, World!\"", text)
	}

	cell, err := screen.GetCell(0, 0)
	if err != nil {
		t.Error(err)
	}

	if cell.String() != "H" {
		t.Errorf("char: \"%s\" is not equal to \"H\"", cell)
	}
}
