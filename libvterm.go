package libvterm

/*
  #cgo pkg-config: vterm
	#include <vterm.h>
	#include "vterm_callbacks.h"
*/
import "C"
import (
	// "log"
	"fmt"
	"unsafe"
)

type VTerm struct {
	vt *C.VTerm
}

type CellP *C.cVTermScreenCell

type Screen struct {
	screen    *C.VTermScreen
	callbacks Callbacks
}

type Callbacks struct {
	sbPushLineCallback func(int, CellP) error
}

func New(rows, cols int) *VTerm {
	vt := C.vterm_new(C.int(rows), C.int(cols))
	C.vterm_set_utf8(vt, 1)
	return &VTerm{vt}
}

func (vt *VTerm) ObtainScreen() *Screen {
	screen := C.vterm_obtain_screen(vt.vt)
	return &Screen{screen: screen}
}

func (vt *VTerm) Close() error {
	_, err := C.vterm_free(vt.vt)
	return err
}

func (vt *VTerm) Write(p []byte) (int, error) {
	l, err := C.vterm_input_write(vt.vt, (*C.char)(unsafe.Pointer(&p[0])), C.size_t(len(p)))
	return int(l), err
}

func (screen *Screen) Reset() error {
	_, err := C.vterm_screen_reset(screen.screen, 1)
	return err
}

func (screen *Screen) SetSbPushLineCallback(callback func(cols int, cells CellP) error) error {
	screen.callbacks.sbPushLineCallback = callback
	ptr := uintptr(unsafe.Pointer(screen))
	_, err := C.screen_set_go_callbacks(screen.screen, C.uintptr_t(ptr))
	return err
}

type Rect struct {
	StartRow int
	EndRow   int
	StartCol int
	EndCol   int
}

func (rect Rect) getVTermRect() C.VTermRect {
	return C.VTermRect{
		start_row: C.int(rect.StartRow),
		end_row:   C.int(rect.EndRow),
		start_col: C.int(rect.StartCol),
		end_col:   C.int(rect.EndCol),
	}
}

func (screen *Screen) GetText(rect Rect) string {
	size := 2 * (rect.EndRow - rect.StartRow) * (rect.EndCol - rect.StartCol)
	buf := make([]byte, 1, size)
	pos := C.vterm_screen_get_text(screen.screen, (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(size), rect.getVTermRect())
	if pos > C.size_t(size) {
		// retry
		buf = make([]byte, 1, pos)
		pos = C.vterm_screen_get_text(screen.screen, (*C.char)(unsafe.Pointer(&buf[0])), pos, rect.getVTermRect())
	}
	return string(buf[0:pos])
}

func (screen *Screen) GetCell(row, col int) (Cell, error) {
	cell := Cell{&C.VTermScreenCell{}}
	ok, err := C.vterm_screen_get_cell(
		screen.screen,
		C.VTermPos{row: C.int(row), col: C.int(col)},
		cell.cell,
	)
	if err != nil {
		return cell, err
	}
	if ok == 0 {
		return cell, fmt.Errorf("Cell(%d, %d) was not found", row, col)
	}
	return cell, nil
}

type Cell struct{ cell *C.VTermScreenCell }

func (cell Cell) String() string {
	var runes []rune
	for i := 0; i < C.VTERM_MAX_CHARS_PER_CELL; i++ {
		r := rune(cell.cell.chars[i])
		if r > 0 {
			runes = append(runes, r)
		}
	}
	return string(runes)
}

//export go_sb_pushline_callback
func go_sb_pushline_callback(cols C.int, cells *C.cVTermScreenCell, user unsafe.Pointer) C.int {
	screen := (*Screen)(user)
	if err := screen.callbacks.sbPushLineCallback(int(cols), CellP(cells)); err != nil {
		return 0
	}
	return 1
}
