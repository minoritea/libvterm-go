#include <vterm.h>
#include "_cgo_export.h"
static VTermScreenCallbacks go_callbacks = {};

void screen_set_go_callbacks(VTermScreen *screen, int type, uintptr_t go_screen) {
	switch(type) {
		case DAMAGE_CALLBACK:
			go_callbacks.damage = &go_screen_damage_callback;
		case SB_PUSHLINE_CALLBACK:
			go_callbacks.sb_pushline = &go_screen_sb_pushline_callback;
	}
	vterm_screen_set_callbacks(screen, &go_callbacks, (void *)go_screen);
}

void go_vterm_output_callback_wrapper(const char *s, size_t len, void *user) {
	go_vterm_output_callback((void *)s, len, user);
}

void vterm_set_go_output_callback(VTerm *vt, uintptr_t go_vterm) {
	vterm_output_set_callback(vt, &go_vterm_output_callback_wrapper, (void *)go_vterm);
}

void convert_vterm_color_to_rgb(VTermColor color, RGB *rgb) {
	rgb->R = color.rgb.red;
	rgb->G = color.rgb.green;
	rgb->B = color.rgb.blue;
}

void unpack_bitfield_vterm_screen_cell_attrs(VTermScreenCellAttrs attrs, UnpackedAttrs *unpacked) {
	unpacked->bold = attrs.bold;
	unpacked->underline = attrs.underline;
	unpacked->italic = attrs.italic;
	unpacked->blink = attrs.blink;
	unpacked->reverse = attrs.reverse;
	unpacked->strike = attrs.strike;
	unpacked->font = attrs.font;
	unpacked->dwl = attrs.dwl;
	unpacked->dhl = attrs.dhl;
}
