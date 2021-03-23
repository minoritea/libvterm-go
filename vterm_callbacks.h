#include <vterm.h>
typedef const VTermScreenCell cVTermScreenCell;
void screen_set_go_callbacks(VTermScreen *screen, int type, uintptr_t go_screen);
void vterm_set_go_output_callback(VTerm *vt, uintptr_t go_vterm);

enum SCREEN_CALLBACK_TYPES {
	DAMAGE_CALLBACK,
	SB_PUSHLINE_CALLBACK,
};

typedef struct {
 uint8_t	R;
 uint8_t	G;
 uint8_t	B;
} RGB;

void convert_vterm_color_to_rgb(VTermColor color, RGB *rgb);

typedef struct {
	uint8_t bold;
	uint8_t underline;
	uint8_t italic;
	uint8_t blink;
	uint8_t reverse;
	uint8_t strike;
	uint8_t font; /* 0 to 9 */
	uint8_t dwl; /* On a DECDWL or DECDHL line */
	uint8_t dhl; /* On a DECDHL line (1=top 2=bottom) */
} UnpackedAttrs;

void unpack_bitfield_vterm_screen_cell_attrs(VTermScreenCellAttrs attrs, UnpackedAttrs *unpacked);
