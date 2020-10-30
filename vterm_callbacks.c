#include <vterm.h>
#include "_cgo_export.h"

static int sb_pushline_callback(int cols, const VTermScreenCell *cells, void *user) {
	return go_sb_pushline_callback(cols, cells, user);
}

static VTermScreenCallbacks go_callbacks = {
	.sb_pushline = &sb_pushline_callback,
};

void screen_set_go_callbacks(VTermScreen *screen, uintptr_t go_screen) {
	vterm_screen_set_callbacks(screen, &go_callbacks, (void *)go_screen);
}
