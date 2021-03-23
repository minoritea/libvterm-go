package libvterm

/*
  #cgo pkg-config: vterm
	#include <vterm.h>
	#include "vterm_callbacks.h"
*/
import "C"
import (
	"fmt"
	"log"
	//	"sync"
	"reflect"
	"unsafe"
)

type VTerm struct {
	vt             *C.VTerm
	outputCallback func([]byte)
	screen         *Screen
	state          *State
	//	sync.Mutex
}

type CellP = *C.cVTermScreenCell

type Screen struct {
	screen    *C.VTermScreen
	callbacks ScreenCallbacks
	scrollbuf [][]*Cell
}

type ScreenCallbacks struct {
	damage      func(rect Rect) error
	moveRect    func(*Screen, *Cell, *Cell) error
	moveCursor  func(*Screen, int, int, int, int, int) error
	setTermProp func(*Screen, unsafe.Pointer, unsafe.Pointer) error
	bell        func(*Screen) error
	resize      func(*Screen, int, int) error
	sbPushLine  func(*Screen, []*Cell) error
	sbPopLine   func(*Screen, []*Cell) error
}

type State struct {
	state *C.VTermState
}

func (state *State) Reset(hard bool) error {
	hardflg := C.int(0)
	if hard {
		hardflg = C.int(1)
	}
	_, err := C.vterm_state_reset(state.state, hardflg)
	return err
}

func New(rows, cols int) *VTerm {
	vt := &VTerm{vt: C.vterm_new(C.int(rows), C.int(cols))}
	vt.ObtainState().Reset(true) // init state
	vt.ObtainScreen().Reset()    // init screen
	C.vterm_set_utf8(vt.vt, 1)
	return vt
}

func (vt *VTerm) ObtainState() *State {
	if vt.state == nil {
		vt.state = &State{state: C.vterm_obtain_state(vt.vt)}
	}
	return vt.state
}

func (vt *VTerm) ObtainScreen() *Screen {
	if vt.screen == nil {
		vt.screen = &Screen{screen: C.vterm_obtain_screen(vt.vt)}
		vt.screen.scrollbuf = make([][]*Cell, 0)
		vt.screen.SetSbPushLineCallback(defaultSbPushLineCallback)
	}
	return vt.screen
}

func (vt *VTerm) Close() error {
	_, err := C.vterm_free(vt.vt)
	return err
}

func (vt *VTerm) Write(p []byte) (int, error) {
	l, err := C.vterm_input_write(vt.vt, (*C.char)(unsafe.Pointer(&p[0])), C.size_t(len(p)))
	return int(l), err
}

func (vt *VTerm) SetOutputCallback(callback func([]byte)) {
	vt.outputCallback = callback
	C.vterm_set_go_output_callback(vt.vt, C.uintptr_t(uintptr(unsafe.Pointer(vt))))
}

func (vt *VTerm) SendUnichar(r rune /*, mod Mod*/) {
	C.vterm_keyboard_unichar(vt.vt, C.uint32_t(r), C.VTERM_MOD_NONE)
}

type Key uint

const (
	KEY_NONE Key = iota
	KEY_ENTER
	KEY_TAB
	KEY_BACKSPACE
	KEY_ESCAPE
	KEY_UP
	KEY_DOWN
	KEY_LEFT
	KEY_RIGHT
	KEY_INS
	KEY_DEL
	KEY_HOME
	KEY_END
	KEY_PAGEUP
	KEY_PAGEDOWN
)

func (vt *VTerm) SendSpecialKey(key Key) {
	/*
		log.Println("vt is locked for SendSpecialKey")
		vt.Lock()
		defer vt.Unlock()
		defer log.Println("vt is unlocked for SendSpecialKey")
	*/
	C.vterm_keyboard_key(vt.vt, C.VTermKey(key), C.VTERM_MOD_NONE)
}

func (screen *Screen) Reset() error {
	_, err := C.vterm_screen_reset(screen.screen, 1)
	return err
}

func defaultSbPushLineCallback(screen *Screen, cells []*Cell) error {
	screen.scrollbuf = append(screen.scrollbuf, cells)
	log.Printf("append new line to scroll buffer(new size: %d)", len(screen.scrollbuf))
	return nil
}

func (screen *Screen) SetDamageCallback(callback func(rect Rect) error) error {
	screen.callbacks.damage = callback
	ptr := uintptr(unsafe.Pointer(screen))
	_, err := C.screen_set_go_callbacks(screen.screen, C.DAMAGE_CALLBACK, C.uintptr_t(ptr))
	return err
}

func (screen *Screen) SetSbPushLineCallback(callback func(*Screen, []*Cell) error) error {
	screen.callbacks.sbPushLine = callback
	ptr := uintptr(unsafe.Pointer(screen))
	_, err := C.screen_set_go_callbacks(screen.screen, C.SB_PUSHLINE_CALLBACK, C.uintptr_t(ptr))
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

func rectFromVTermRect(rect C.VTermRect) Rect {
	return Rect{
		StartRow: int(rect.start_row),
		EndRow:   int(rect.end_row),
		StartCol: int(rect.start_col),
		EndCol:   int(rect.end_col),
	}
}

func (screen *Screen) GetText(rect Rect) string {
	size := 2*(rect.EndRow-rect.StartRow)*(rect.EndCol-rect.StartCol) + 1
	buf := make([]byte, 1, size)
	pos := C.vterm_screen_get_text(screen.screen, (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(size), rect.getVTermRect())
	if pos > C.size_t(size) {
		// retry
		buf = make([]byte, 1, pos)
		pos = C.vterm_screen_get_text(screen.screen, (*C.char)(unsafe.Pointer(&buf[0])), pos, rect.getVTermRect())
	}
	return string(buf[:pos])
}

type Color struct {
	R uint8 `json:"red"`
	G uint8 `json:"green"`
	B uint8 `json:"blue"`
}

type CellAttrs struct {
	Bold      uint8 `json:"bold"`
	Underline uint8 `json:"underline"`
	Italic    uint8 `json:"italic"`
	Blink     uint8 `json:"blink"`
	Reverse   uint8 `json:"reverse"`
	Strike    uint8 `json:"strike"`
	Font      uint8 `json:"font"`
	DWL       uint8 `json:"dwl"`
	DHL       uint8 `json:"dhl"`
}

type Cell struct {
	cell    *C.VTermScreenCell
	Attrs   CellAttrs `json:"attrs"`
	FGColor Color     `json:"fg_color"`
	BGColor Color     `json:"bg_color"`
	Text    string    `json:"text"`
}

func convertScreenCellToCell(screen *Screen, cellp CellP) *Cell {
	cell := &Cell{}
	cell.cell = cellp
	cellattrs := C.UnpackedAttrs{}
	C.unpack_bitfield_vterm_screen_cell_attrs(cellp.attrs, &cellattrs)
	cell.Attrs.Bold = uint8(cellattrs.bold)
	cell.Attrs.Underline = uint8(cellattrs.underline)
	cell.Attrs.Italic = uint8(cellattrs.italic)
	cell.Attrs.Blink = uint8(cellattrs.blink)
	cell.Attrs.Reverse = uint8(cellattrs.reverse)
	cell.Attrs.Strike = uint8(cellattrs.strike)
	cell.Attrs.Font = uint8(cellattrs.font)
	cell.Attrs.DWL = uint8(cellattrs.dwl)
	cell.Attrs.DHL = uint8(cellattrs.dhl)
	var rgb C.RGB
	C.vterm_screen_convert_color_to_rgb(screen.screen, &cellp.fg)
	C.convert_vterm_color_to_rgb(cellp.fg, &rgb)
	cell.FGColor.R = uint8(rgb.R)
	cell.FGColor.G = uint8(rgb.G)
	cell.FGColor.B = uint8(rgb.B)
	C.vterm_screen_convert_color_to_rgb(screen.screen, &cellp.bg)
	C.convert_vterm_color_to_rgb(cellp.bg, &rgb)
	cell.BGColor.R = uint8(rgb.R)
	cell.BGColor.G = uint8(rgb.G)
	cell.BGColor.B = uint8(rgb.B)
	var runes []rune
	for i := 0; i < C.VTERM_MAX_CHARS_PER_CELL; i++ {
		r := rune(cellp.chars[i])
		if r == 0 {
			break
		}
		runes = append(runes, r)
	}
	cell.Text = string(runes)
	return cell
}

func (screen *Screen) FetchCell(row, col int) (*Cell, error) {
	if row >= 0 {
		return screen.GetCell(row, col)
	}

	log.Println("minus row", row)

	// fetch from scroll buffer

	index := len(screen.scrollbuf) + row
	if index < 0 || index > len(screen.scrollbuf) {
		return nil, fmt.Errorf("Out of range access failed(row: %d)", row)
	}
	line := screen.scrollbuf[len(screen.scrollbuf)+row]

	if len(line) > col {
		return line[col], nil
	}

	// return empty cell
	var cell Cell
	cell.Text = ""
	cell.BGColor = line[len(line)-1].BGColor
	return &cell, nil
}

func (screen *Screen) GetCell(row, col int) (*Cell, error) {
	cellp := &C.VTermScreenCell{}
	ok, err := C.vterm_screen_get_cell(
		screen.screen,
		C.VTermPos{row: C.int(row), col: C.int(col)},
		cellp,
	)
	if err != nil {
		return nil, err
	}
	if ok == 0 {
		return nil, fmt.Errorf("Cell(%d, %d) was not found", row, col)
	}
	return convertScreenCellToCell(screen, cellp), nil
}

func (cell Cell) String() string {
	return cell.Text
}

//export go_screen_sb_pushline_callback
func go_screen_sb_pushline_callback(cols C.int, cells *C.cVTermScreenCell, user unsafe.Pointer) C.int {
	var sliceCells []C.cVTermScreenCell
	header := (*reflect.SliceHeader)(unsafe.Pointer(&sliceCells))
	header.Data = uintptr(unsafe.Pointer(cells))
	header.Len = int(cols)
	header.Cap = int(cols)
	screen := (*Screen)(user)
	var goCells []*Cell
	for _, cell := range sliceCells {
		goCells = append(goCells, convertScreenCellToCell(screen, &cell))
	}
	if err := screen.callbacks.sbPushLine(screen, goCells); err != nil {
		return 0
	}
	return 1
}

//export go_screen_damage_callback
func go_screen_damage_callback(rect C.VTermRect, user unsafe.Pointer) C.int {
	screen := (*Screen)(user)
	if err := screen.callbacks.damage(rectFromVTermRect(rect)); err != nil {
		return 0
	}
	return 1
}

//export go_vterm_output_callback
func go_vterm_output_callback(bytes unsafe.Pointer, size C.size_t, user unsafe.Pointer) {
	vt := (*VTerm)(user)
	vt.outputCallback(C.GoBytes(bytes, C.int(size)))
}
