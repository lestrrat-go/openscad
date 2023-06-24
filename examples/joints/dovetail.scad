function dovetail_backtrack_height(base) = base*sqrt(2);
function dovetail_top_width(base, height, backtrack_height) = base*((backtrack_height+height)/backtrack_height);
function dovetail_top_offset(base, height) = 
    let(
        backtrack_height=dovetail_backtrack_height(base),
        top_width=dovetail_top_width(base, height, backtrack_height)
    )
        (top_width-base)/2;

module dovetail_tenon(base, height, depth)
{
	// calculate starting from the "bottom" (base)of the ari
	//
	//   top
	//  ____
	//  \__/
	//   bottom
	//
	// we first find the point perpendicular from the middle
	// of the bottom line, with length bottom*sqrt(2)
    xoffset=dovetail_top_offset(base, height);
    linear_extrude(height=depth)
        polygon(points=[[xoffset, 0], [xoffset+base, 0], [xoffset*2+base, height], [0, height]]);
}