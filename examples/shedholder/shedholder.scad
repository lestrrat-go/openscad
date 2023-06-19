
include <threads.scad>
include <constants.scad>

// The total depth (well, height) of the piece, including
// the ring, the base, and the leg. 
outer_depth=9*cm;

// Dimensions of the piece of wood that would be inserted
// in the ring.
wood_height=8.9*cm;
wood_depth=1.9*cm;

separator_column_width=0.5*cm;
ring_width=5*cm;
ring_corner_radius=0.75*cm;

outer_height=wood_height*2 + separator_column_width + ring_corner_radius *2;
inner_height=wood_height*2 + separator_column_width + leeway*2;


rod_radius=1.5*cm;

skirt_rod_radius=rod_radius+0.5*cm;

leg_length=outer_depth-(wood_depth+ring_corner_radius*2+15);
notch_dim=rod_radius*3/4;
ari_base=rod_radius/5;

inch=2*cm;

function dovetail_backtrack_height(base) = (base*sqrt(2));
function dovetail_top_width(base, height, backtrack_height) = (base*((backtrack_height+height)/backtrack_height));
function dovetail_top_offset(base, height) = ((dovetail_top_width(base, height, dovetail_backtrack_height(base))-base)/2);

module ari(base, height, depth) {
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
    polygon(points=[[xoffset, 0], [(xoffset+base), 0], [((xoffset*2)+base), height], [0, height]]);
    
}

module nbym( width, height, depth, leeway=[0, 0, 0]) {
    cube([height*inch, width*inch, depth]+leeway);
}

// The standard size in this design (could be 1x4, 2,4, etc)
module wood_piece(width, leeway=[0, 0, 0]) {
    cube([width, wood_height, wood_depth]+leeway);
}


module ring() {    
    render()
    {
        // two rows of dovetails
        for(xoffset=[-rod_radius*3/5, rod_radius/5], yoffset=[rod_radius*3, outer_height-ring_corner_radius*2+leeway*2]) {
            translate([ring_width/2+rod_radius/2+xoffset-ari_base/2, yoffset, -ring_corner_radius/2])
                rotate([90, 180, 0])
                ari(ari_base, ring_corner_radius/2, rod_radius*3);
        }

        difference()
        {
            hull()
            {
                for(yloc=[0, wood_height*2+separator_column_width], zloc=[0, wood_depth]) {
                    translate([ring_width, yloc, zloc])
                        rotate([0, -90, 0])
                        cylinder(h=ring_width, r=ring_corner_radius, $fn=100);
                }
            }
            
            // punch holes the size of the wood
            for (yloc=[0, wood_height+separator_column_width]) {
                translate([0, yloc+leeway, 0.25*cm])
                    wood_piece(ring_width, leeway=[0, leeway,leeway]);
            }
            
            // punch holes for the screw rod base, and create shallow hole to inser the leg

            hull() {
                for(yloc=[ring_corner_radius*2, inner_height-ring_corner_radius*2]) {
                        translate([ring_width/2, yloc, -ring_corner_radius])
                            cylinder(r=rod_radius+leeway, h=ring_corner_radius/2, $fn=400);

                }
            }
        }
    }
}

module holder_base() {
    translate([0, -5, outer_depth-outer_depth/3-1.7/*-ring_corner_radius*2*/]) {
        //render() difference()
        {
        ring();
            //translate([0, -10, -1]) cube([ring_width, outer_height*2, outer_depth]);
            //translate([0, outer_height/2, -10]) cube([ring_width, outer_height, outer_depth]);
        }
    }
}

module holder_leg_base() {
    render() 
        difference()
    {
        rotate([0, 180, 0])
            RodStart(rod_radius*2, 10);
                rotate([0, 180, 0])

        // two rows of dovetails
        for(xoffset=[-rod_radius*3/5, rod_radius/5]) {
            translate([xoffset, rod_radius, 0])
            rotate([90, 0, 0])
                ari(ari_base+leeway, ring_corner_radius/2+leeway, rod_radius*2);
        }
    }
}

module holder_leg_bases() {
    translate([0, 0, outer_depth-(wood_depth+ring_corner_radius*2)]) {
        for(yloc=[ring_corner_radius*2, inner_height-ring_corner_radius*2]) {
            translate([ring_width/2, yloc, 0])
                holder_leg_base();
        }
    }
}

module holder_leg() {
    RodEnd(rod_radius*2, leg_length);
}

module holder_legs() {
    for(yloc=[ring_corner_radius*2, inner_height-ring_corner_radius*2]) {
        translate([ring_width/2, yloc, 0])
            holder_leg();
    }
}

module skirt() {
    module skirt_rod() {
    render()
    {
        difference()
        {
            cylinder(r=skirt_rod_radius, h=20, $fn=400);
            cylinder(r=rod_radius+leeway, h=20, $fn=400);
        }
    }
}

    for(yloc=[ring_corner_radius*2, inner_height-ring_corner_radius*2]) {
        translate([ring_width/2, yloc, 0])
            skirt_rod();
    }
    translate([0, -ring_corner_radius, 0])
        cube([ring_width, outer_height, 1]);
}

{
    color("White", 1) translate([0, 0, 20]) holder_base();

    // Leg bases (screw rod base)
    // a) Use this to check the design
    color("Red", 1) holder_leg_bases();
    // b) Use this when printing
    // color("Red", 1) holder_leg_base();

    // Legs (screw rod legs)
    // a) Use this to check the design
    color("Green", 1) holder_legs();
    // b) Use this when printing
    // color("Green", 1) holder_leg();
    
    color("Blue", 1) translate([0, 0, -30]) skirt();
}

