include <dovetail.scad>
include <threads.scad>
include <gears.scad>

module gear_toy(
    print_mode=true,
    modul=1,
    gear_teeth=30,
    pinion_teeth=11,
    base_thickness=3, // thickness of the bottom place (gear_base)
    rod_length=7,
    helix_angle=20,
    pressure_angle=20,
    gear_box_base_thickness=2,
    gear_box_wall_thickness=2,
    gear_box_floor_thickness=2,
    gear_box_shaft_height=2 ,
    gear_box_bottom_gear_clamp_height=3,
    gear_shaft_height=2,
    spur_gear_bore=6,
    spur_gear_width=5,
    cradle_thickness=3,
    mid_bar_length=20,
    axis_angle=90,
    tooth_width=5,
    crank_holder_length=5,
    top_gear_clamp_height=2,
    stopper_dim = 2,
    crank_handle_radius=2,
    crank_handle_height=2,
    crank_handle_length=25,
    bottom_gear_clamp_height=3
) {
    r_hole = sg_weight_reduction_hole_radius(modul, gear_teeth, spur_gear_bore); 
    rm = sg_weight_reduction_hole_distance(modul, gear_teeth, spur_gear_bore);
    z_hole = sg_weight_reduction_hole_count(modul, gear_teeth, spur_gear_bore);
    cradle_radius=r_hole+rm+1;
    height_f_gear = bgp_cone_height(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion=false);
    height_f_pinion = bgp_cone_height(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion=true);
    delta_gear = bgp_gear_cone_angle(axis_angle, gear_teeth, pinion_teeth);
    delta_pinion = bgp_pinion_cone_angle(axis_angle, gear_teeth, pinion_teeth);

module place(design_xlate=[0, 0, 0], design_rotate=[0, 0, 0], print_xlate=[0, 0, 0], print_rotate=[0, 0, 0]) {
    translate(print_mode ? print_xlate : design_xlate)
        rotate(print_mode ? print_rotate : design_rotate)
            children();
}

// gear_base_radius returns the radius that can be used to draw the
// hexagon that the gear teeth will be drawn on.  This is the NOT
// the same as the pitch circle radius.
//
// what we want is the radius `r2` that would allow `cylinder` to draw
// a hexagon which produces `h` below.
//
// calculations:
//    `h` is a point in the circle drawn using radius `r`. This is
//    where the gear teeth will mesh with its counter part. so h = r.
//
//    then, cos(30) =  r2/r, so r2 = r * cos(30)
//    ______________
//   /  ____|_____/ \
//    /  \  | h /  \  \
//   /  r \ |  / r2 \
//  /      \. /      \
//  \                /
//   \              /
//    \            /
//     \__________/
function gear_base_radius(modul, gear_teeth) =
    let (r = sg_pitch_circle_radius(modul, gear_teeth))
        r/cos(30);

// This bevel gear has a knob on the top side to insert a shaft
// with notches to drive the gear or the shaft.
module knobbed_bevel_gear(modul, gear_teeth, delta, tooth_width, gear_bore, pressure_angle, helix_angle,
    knob_radius = 2.5,
    knob_width = 2,
    knob_height = 2
) {
    bevel_gear_height=bevel_gear_cylinder_height(modul, gear_teeth, delta, tooth_width);

    // the knob is a washer with a few notches on the side
    // to drive the shaft
    translate([0, 0, bevel_gear_height]) {
        render() difference() {
            washer(gear_bore/2+2, gear_bore/2, knob_height);
            for (i=[0:3]) {
                rotate([0, 0, 90*i]) {
                    translate([gear_bore/2*0.95, -gear_bore/4, 0])
                    cube([3, 2, knob_height]);
                }
            }
        }
    }

    // the gear itself
    bevel_gear(modul, gear_teeth, delta, tooth_width, gear_bore, pressure_angle, helix_angle);
}

module bevel_gear_bottom_clamp(gear_bore, bottom_gear_clamp_height) {
    render() difference() {
        washer(gear_bore*2, gear_bore/2, bottom_gear_clamp_height);
        translate([-1.01, -gear_bore*2, 0]) {
            cube([2, gear_bore*4, 2]+[0.2, 0.6, 0.2]);
        }
    }
}

module bevel_gear_floor(inner_radius, gear_bore, wall_thickness, floor_thickness, clearance=0.3) {
    shaft_radius=gear_bore/2; // radius of shaft itself, does not include wall thickness
    // floor to house the shaft
    render() difference() {
       cylinder(r=crank_handle_radius+4, h=floor_thickness-clearance, $fn=6);

        // same size hole as the shaft needs to be punched open
        // in the floor of the box 
        cylinder(r=shaft_radius, h=gear_shaft_height, $fn=400);
    }
    render() difference() {
        // outer hexagonal cylinder. this stays. everything below is something
        // that we chip out of this object.
        cylinder(r=inner_radius-clearance, h=floor_thickness-clearance, $fn=6);

        // get rid of the walls towards the inside
        translate([-(inner_radius+wall_thickness), 0, 0])
            cube([2*(inner_radius+wall_thickness), inner_radius, inner_radius+wall_thickness*2]);

        // same size hole as the shaft needs to be punched open
        // in the floor of the box
        cylinder(r=shaft_radius, h=gear_shaft_height, $fn=400);
    }
}

module bevel_pinion() {
    pinion_bore=(crank_handle_radius+0.15)*2;
    bevel_gear(modul, pinion_teeth, delta_pinion, tooth_width, pinion_bore, pressure_angle, helix_angle);
    translate([0.3, -(pinion_bore/2+0.3), 0])
        cube([pinion_bore/2,pinion_bore/2, bevel_gear_cylinder_height(modul, pinion_teeth, delta_pinion, tooth_width)]);
}

module bevel_gear_washer(gear_bore) {
    shaft_radius=gear_bore/2; // radius of shaft itself, does not include wall thickness
    let (washer_height=gear_shaft_height+bottom_gear_clamp_height) {
        #render() difference() {
            washer(shaft_radius+1, shaft_radius, washer_height);
            translate([0, -shaft_radius*3/2, washer_height/2])
                rotate([-90, 0, 0])
                    cylinder(r=1.1, h=shaft_radius*3, $fn=100);
        }
    }
}


// gear_box draws a box that can be used to house a bevel gear pair
module gear_box(modul, gear_teeth, gear_bore, floor_thickness=2, wall_thickness=2,
) {
    // because the gear box requies that it aligns wit the bevel gears,
    // we are going to draw them within this coordinate system to make
    // the calculations easier
    axis_angle=90;
    pinion_teeth=11;
    pinion_bore=4;
    shaft_radius=gear_bore/2; // radius of shaft itself, does not include wall thickness


    if (!print_mode) {
        // main gear. The gear will be secured using two pieces
        // of rods that act as clamps
        translate([0, 0, floor_thickness+gear_shaft_height+0.2+bottom_gear_clamp_height])
            knobbed_bevel_gear(modul, gear_teeth, delta_gear, tooth_width, gear_bore, pressure_angle, helix_angle);

        // bottom gear clamp
        translate([0, 0, gear_shaft_height]) {
            *translate([0, 0, floor_thickness]) {
                bevel_gear_bottom_clamp(gear_bore, bottom_gear_clamp_height);
            }

            bevel_gear_washer(gear_bore);
        }
    }

    height_f_pinion = bgp_cone_height(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion=true);
    height_f_gear = bgp_cone_height(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion=false);

    gear_blade_clearance=2;
    r=crank_wall_inner_radius(modul, gear_teeth, crank_handle_washer_height, gear_blade_clearance);

    if (!print_mode) {
        // pinion
        translate([0, -height_f_pinion*cos(90-axis_angle),height_f_gear-height_f_pinion*sin(90-axis_angle)+floor_thickness+gear_shaft_height+bottom_gear_clamp_height])
            rotate([-axis_angle,axis_angle,0])
                bevel_pinion();

        // washer to hold the pinion.
        translate([0,
            -r+wall_thickness+crank_handle_height/2+2-0.26,
            height_f_gear-height_f_pinion*sin(90-axis_angle)+floor_thickness+crank_handle_height+bottom_gear_clamp_height
        ])  
        rotate([-axis_angle, axis_angle])
            washer(crank_handle_radius+2, crank_handle_radius, crank_handle_height);

        // floor
//        bevel_gear_floor(r, gear_bore, wall_thickness, floor_thickness);
    }
}
    
module gear_base(modul, gear_teeth, bore, base_thickness=3, male=true) {
    r=gear_base_radius(modul, gear_teeth);
    let($fn=400) {
        if (male) {
            translate([0, 0, base_thickness]) {
                cylinder(r=bore/2-0.2, h=10);
                cylinder(r=5, h=2);
            }
        } else {
            translate([0, 0, base_thickness]) {
                difference() {
                    cylinder(r=5, h=3);
                    cylinder(r=bore/2, h=10);
                }
            }
        }
    }
    
    render() difference() {
        union() {
            cylinder(r=r, h=base_thickness, $fn=6);  
            
            // tenons
            for (zangle=[60, 180, 300]) {
                rotate([0, 0, zangle])
                    translate([-dovetail_top_width(5, 5, dovetail_backtrack_height(5))/2, cos(30)*r-0.2, 0]) dovetail_tenon(5, 5, 3);
            }
        }         
        // mortices
        for (zangle=[0, 120, 240]) {
            rotate([0, 0, zangle])
                translate([dovetail_top_width(5, 5, dovetail_backtrack_height(5))/2, cos(30)*r+0.2, 0])
                //translate([0, 5, 0])
                rotate([0, 0, 180])
                dovetail_tenon(5.1, 5.1, 3);
        }                           
    }
}

function crank_wall_inner_normal(modul, gear_teeth, washer_height, clearance=2) =
        // radius of the gear
        sg_tip_circle_radius(modul, gear_teeth)+
        // height of the crank handle shaft
        washer_height+
        // We need a bit of clearance
        clearance;

function crank_wall_inner_radius(modul, gear, washer_height, clearance) =
    let (
        inner_normal=crank_wall_inner_normal(modul, gear, washer_height, clearance)
    )
        inner_normal/cos(30);

function crank_wall_outer_normal(modul, gear, washer_height, wall_thickness, clearance) =
    let (
        inner_normal=crank_wall_inner_normal(modul, gear, washer_height, clearance)
    )
        inner_normal+wall_thickness;

function crank_wall_outer_radius(modul, gear, washer_height, wall_thickness=2, clearance=2) =
    let (
        inner_radius=crank_wall_inner_radius(modul, gear, washer_height, clearance),
        inner_normal=crank_wall_inner_normal(modul, gear, washer_height, clearance),
        outer_normal=crank_wall_outer_normal(modul, gear, washer_height, wall_thickness, clearance),
        ratio=outer_normal/inner_normal
    )
        inner_radius*ratio;

// TODO: move it around
crank_handle_washer_height=2;


module crank_handle_shaft_holder(thickness=2, blade_clearance=2) {
    crank_handle_shaft_radius=2.1;
    wall_outer_radius=crank_wall_outer_radius(modul, gear_teeth, crank_handle_washer_height, thickness, blade_clearance);
    // crank handle shaft casing
    color("LightGreen", 1) {
        washer(4, crank_handle_shaft_radius, crank_holder_length+4, outer_fn=6);
        washer(6, crank_handle_shaft_radius, crank_holder_length);
        translate([0, 0, crank_holder_length+2.2])
            washer(6, crank_handle_shaft_radius, 2);
    }
}

module crank_wall(r, crank_zloc, gear_box_floor_zloc, thickness=2, blade_clearance=2) {
    // this radius must be able to hold both the gears and the small watch that
    // ohuses the crank handle shaft
    wall_inner_radius=crank_wall_inner_radius(modul, gear_teeth, crank_handle_washer_height, blade_clearance);
    wall_inner_normal=crank_wall_inner_normal(modul, gear_teeth, crank_handle_washer_height, blade_clearance);
    wall_outer_normal=crank_wall_outer_normal(modul, gear_teeth, crank_handle_washer_height, thickness, blade_clearance);

    //hex_inner_edge_distance=floor((r/2)/tan(30))+blade_clearance- /*manual adjustment*/0.1;
    //hex_outer_edge_distance=hex_inner_edge_distance+thickness;
    //inner_ratio=hex_inner_edge_distance/((r/2)/tan(30));
    //outer_ratio=hex_outer_edge_distance/((r/2)/tan(30));
    //// wall_inner_radius=r*inner_ratio;
    //wall_outer_radius=r*outer_ratio;
    wall_outer_radius=crank_wall_outer_radius(modul, gear_teeth, crank_handle_washer_height, thickness, blade_clearance);
    height=base_thickness*2+cradle_thickness+spur_gear_width+mid_bar_length+gear_box_base_thickness;
    
    // A wall to hold the crank handle shaft. its width 
    let(column_radius=3, zloc_shaft_center=gear_box_shaft_height+gear_box_bottom_gear_clamp_height+height_f_gear+0.3) {
        // We want a wall that has width=wall_inner_radius-4 (2 from left and right),
        // placed at the center of the wall
        for (i=[-1, 1]) {
            translate([
                i*(wall_outer_radius/2-(column_radius)),
                -(wall_outer_normal-(column_radius))+2,
                2
            ]) {
                // This is the column to house the other end of the wall
                // that holds the crank handle shaft in place
                translate([0, -column_radius*2, gear_box_floor_zloc]) {
                    RodStart(column_radius*2-1, gear_box_shaft_height+gear_box_bottom_gear_clamp_height+height_f_gear*2-1, thread_len=4);
                }
                //cylinder(r=column_radius, h=gear_box_shaft_height+gear_box_bottom_gear_clamp_height+height_f_gear*2);
                // connect the walls and the columns as a hull
                translate([0, -column_radius*2, -2])
                    rotate([0, 0, 180])
                    hull() {
                        cylinder(r=column_radius, h=height, $fn=100);
                        translate([-column_radius, -column_radius, 0])
                            cube([column_radius*2, 1, height]);
                    }
            }
        }
            
        color("Magenta", 1) place(
            design_xlate=[0, 0, gear_box_floor_zloc+gear_box_floor_thickness], // +zloc_shaft_center+3],
            print_xlate=[0, -20, 0]
        ) {
            for (i=[-1, 1]) {
                // These are the columns to hold the top end of the wall
                translate([
                    i*(wall_outer_radius/2-(column_radius)),
                    -(wall_outer_normal-(column_radius))-column_radius*2+2,
                    0
                ]) {
                    rotate([0, 0, 180])
                        render() difference() {
                            hull() {
                                cylinder(r=column_radius, h=zloc_shaft_center*1.5, $fn=100);
                                translate([-column_radius, -column_radius, 0])
                                    cube([column_radius*2, 1, zloc_shaft_center*1.5]);
                            }
                            cylinder(r=column_radius-0.3, h=zloc_shaft_center*1.5, $fn=100);
                        }
                }
            }
            // This is the wall with a hole in the middle to hold the crank handle
            render() difference()
            {
                union() {
                    translate([
                        -(wall_outer_radius-column_radius*4)/2,
                        -wall_inner_normal-2,
                        0
                    ]) {
                        cube([wall_outer_radius-column_radius*4, 2, zloc_shaft_center*1.5]);
                    }
                    translate([0, -wall_outer_normal, zloc_shaft_center])
                        rotate([90, 0, 0])
                            cylinder(r=4, h=4, $fn=100);
                    }
                    #translate([0, -wall_inner_normal, zloc_shaft_center])
                        rotate([90, 0, 0])
                            cylinder(r=2.1, h=6, $fn=100);
            }
        }

            place(
                design_xlate=[0, 0, gear_box_floor_zloc+zloc_shaft_center/2+3+ zloc_shaft_center*2
            ]) {
                for(i=[-1,1]) {
                    place(
                        print_xlate=[0, -15, 0]
                    )
                    // Rod ends to hold the wall in place
                    translate([
                        i*(wall_outer_radius/2-(column_radius)),
                        -(wall_outer_normal-(column_radius))+2,
                    ]) {
                            translate([0, 0, 2])
                                RodEnd(column_radius*2-1, 8, thread_len=4);
                            render() difference() {
                                cylinder(r=column_radius+1, h=2, $fn=100);
                                translate([-1, -(column_radius+1), 0])
                                    cube([2, (column_radius+1)*2, 1]);
                            }
                    }
                }
            }

        

        // floor to rest the bevel gear
        translate([0, 0, gear_box_floor_zloc])
            bevel_gear_floor(wall_outer_radius, spur_gear_bore, 2, 2, clearance=0);
            difference() {
                union() {
                difference() {
                    // main (outer) hexagonal cylinder
                    cylinder(r=wall_outer_radius, h=height, $fn=6);

                    cylinder(r=r+0.3, h=3, $fn=6);
                    translate([0, 0, 3])
                        cylinder(r=wall_inner_radius, h=height, $fn=6);
                }
    
            
                // tenons on the foot
                for (zangle=[120, 240]) {
                    rotate([0, 0, zangle])
                        translate([dovetail_top_width(5, 5, dovetail_backtrack_height(5))/2, cos(30)*r+0.27, 0])
                        rotate([0, 0, 180])
                        dovetail_tenon(5, 5, 3);
                }

                // mortices for the crank handle shaft casing
                // they will be shaped further in later sections
                /*
                translate([0, -wall_outer_normal, crank_zloc])
                rotate([-90, 0, 0]) {
                    // width = 4 because they house a width 2 (+clearance) tenons
                    let (tenon_width=2, clearance=0.2, edge_width=1, shaft_radius=4) {
                        translate([-(tenon_width+edge_width), -(shaft_radius+2), 2]) 
                            cube([(tenon_width+edge_width)*2, (shaft_radius+2)*2, 2]);
                        translate([-(shaft_radius+2), -(tenon_width+edge_width), 2])
                            cube([(shaft_radius+2)*2, (tenon_width+edge_width)*2, 2]);
                    }
                }
                */
            }
            translate([-wall_outer_radius, 0, 0])
                cube([wall_outer_radius*2, wall_outer_radius, height]);
        
            // mortices
            for (zangle=[180]) {
                    rotate([0, 0, zangle])
                        translate([-dovetail_top_width(5, 5, dovetail_backtrack_height(5))/2, cos(30)*r-0.2, 0]) dovetail_tenon(5.1, 5.1, 3);
            }
            
            // hole on the wall for the crank handle housing
            // the housing is mario-style pipe with a hole the size of
            // of the crank handle shaft + clearance
            translate([0, -wall_outer_normal-2, crank_zloc])
            rotate([-90, 0, 0]) {
                cylinder(r=4+0.2, h=5);  // 2 for the crank handle shaft, 2 for surrounding
                /*
                // this cylinder is used as the stopper so that the
                // housing does not 
                translate([-1.1, -6.1, 0.5])
                    cube([2.2, 12.2, 4]);
                translate([-6.1, -1.1, 0.5])
                    cube([12.2, 2.2, 4]);
                    */
            }
        }
    }
}
    
module shaft_handle() {
    render() difference() {
    // main cylinder (grip)
    RodEnd(10, 40, 10, 6.4*0.75);

    // Create a hole so that we can stick in a screwdriver (or the like)
    // to tighten the threads
    translate([0, -5, 10])
        rotate([-90, 0, 0])
        cylinder(r=2, h=10, $fn=100);
    }
}

module shaft_closer() {
    render() difference() {
        cylinder(r=5, h=2);
        // Add a groove so that we can use screwdrivers, if need be
        translate([-5, -1, 0])
            cube([10, 2, 2]);
    }
    translate([0, 0, 2])
        RodStart(6.4, 5.2, 10);
}

module crank_arm(crank_handle_length=25, endpiece_radius=5, thickness=5) {
    // vertical board
    render() difference() {
        hull() {
            translate([0, 0, crank_handle_length])
                rotate([-90, 0, 0])
                    cylinder(r=5, h=thickness);
            translate([0, 0, 0])
                rotate([-90, 0, 0])
                    cylinder(r=5, h=thickness);
            translate([-endpiece_radius, 0, 0])
                cube([endpiece_radius*2, thickness, crank_handle_length]);
        }
        
        rotate([-90, 0, 0])
            cylinder(r=3.6, h=thickness);
    }
    
    translate([0, 0, crank_handle_length]) {
        rotate([-90, 0, 0]) {
            translate([0, 0, thickness])
                cylinder(r=3, h=2.2);
            render() difference() {
                cylinder(r=2, h=23);
                translate([-0.2, -0.2, 18])
                    cube([2.2, 2.2, 9]);
            }
        }
    }
}

module crank_handle() {
    let ($fn=400) {
        if (print_mode) {
            translate([55, -60, 0])
                rotate([90, 0, 0])
                crank_arm(crank_handle_length=crank_handle_length);

            translate([40, -75, 0])
                shaft_handle();

            translate([40, -90, 0])
                shaft_closer();
        } else {
            *translate([0, -33, 22.6]) {
                crank_arm(crank_handle_length=crank_handle_length);
                translate([0, -43, 0])
                    rotate([-90, 0, 0]) {
                        shaft_handle();
            
                        // holder rod
                        translate([0, 0, 50])
                            rotate([0, 180, 0])
                            shaft_closer();
                    }
            }
        }
    }
}

module drive_shaft_stopper(dim, stopper_length){
    rotate([0, 90, 0])
    render() difference() {
        cube([dim, stopper_length, dim]);
        translate([0, (stopper_length-spur_gear_bore)/2+0.1, -dim/2])
            cube([dim, spur_gear_bore, dim]+[0.1, 0.2, 0.1]);
    }
}

module drive_shaft(shaft_length) {
    delta_gear = bgp_gear_cone_angle(axis_angle, gear_teeth, pinion_teeth);

    zloc_bevel_clamp=base_thickness+cradle_thickness+spur_gear_width+mid_bar_length+gear_box_base_thickness+(gear_box_shaft_height+gear_box_bottom_gear_clamp_height)/2;
    zloc_over_spur_gear=base_thickness+cradle_thickness+spur_gear_width+0.2;
    zloc_cradle_bottom=base_thickness+stopper_dim/2;

    // The actual stoppers. We create them with -0.2 dim 
    *for (zloc=[
        zloc_bevel_clamp,
        zloc_over_spur_gear,
        zloc_cradle_bottom,
    ]) {
        let(dim=stopper_dim-0.2) {
        bar_length=(zloc==zloc_cradle_bottom)?cradle_radius: spur_gear_bore*2;
        color("Red", 1)
        translate([-dim/2, 0 /*-bar_length/2*/ /* centering is prettier, but harder to print */, zloc+dim/2-dim]) {
            drive_shaft_stopper(stopper_dim-0.2, bar_length);
        }
        }
    }

    //  connectors at the top end of the shaft
    color("LightGreen", 1)
    translate([0, 0, base_thickness+cradle_thickness+spur_gear_width+mid_bar_length+gear_box_base_thickness+gear_box_shaft_height+gear_box_bottom_gear_clamp_height+
            bevel_gear_cylinder_height(modul, gear_teeth, delta_gear, tooth_width)+top_gear_clamp_height
    ])
    // translate([0, 0, base_thickness+gear_box_shaft_height+0.2+gear_box_bottom_gear_clamp_height+
    //    bevel_gear_cylinder_height(modul, gear_teeth, delta_gear, tooth_width)]) 
                for (i=[0:3]) {
                    rotate([0, 0, 90*i])
                        translate([(spur_gear_bore/2-0.5), -0.9, -top_gear_clamp_height])
                            cube([2.5, 1.8, top_gear_clamp_height]);
                }

    // holes for the stoppers. we use stopper_dim+0.2 here to get a bit of extra clearance
    let($fn=400, dim=stopper_dim+0.2) {
        color("LightGreen", 1)
        render() difference() {
            cylinder(r=spur_gear_bore/2-0.2, h=shaft_length);
            for (zloc=[
                zloc_bevel_clamp,
                zloc_over_spur_gear,
                zloc_cradle_bottom
            ]) {
                if (zloc==zloc_bevel_clamp) {
                    translate([0, -spur_gear_bore/2, zloc])
                    rotate([-90, 0, 0])
                        cylinder(r=1, h=spur_gear_bore, $fn=100);
                } else {
                    translate([-dim/2, -spur_gear_bore/2, zloc-dim/2]) {
                        cube([dim, spur_gear_bore, dim]);
                    }
                }
            }
        }
    }
}

module spur_gear_cradle() {
    let ($fn=400) {
    translate([0, 0, cradle_thickness]) {
        for (i=[0:z_hole-1]) {
            rotate([0, 0, 360/z_hole*i])
                translate([rm, 0, 0])
                    cylinder(r=r_hole-0.2, h=spur_gear_width+1);
        }
    }
    // cradle_bottom
    render() difference() {
        cylinder(r=cradle_radius, h=cradle_thickness);
        cylinder(r=spur_gear_bore/2+0.2, h=cradle_thickness);
        translate([-1.01, -cradle_radius/2, 0]) {
            cube([2, cradle_radius, 2]+[0.2, 0.6, 0.2]);
        }
    }
    }
}


module all() {
    blade_clearance=2;
    pinion_distance=/*height_f_gear-*/height_f_pinion*sin(90-axis_angle);
    echo("height_f_gear = ", height_f_gear);
    echo("height_f_pinion = ", height_f_pinion);
    echo("pinion_distance = ", pinion_distance);

    module guide(name, height) {
        render() difference() {
            cube([50, 2, height]);
            translate([50, 0, 0])
            rotate([90, 0, 180])
                linear_extrude(2)
                    text(str(name, " (", height, ")"), valign="bottom", size=height > 4 ? 3 : height-1);
        }
    }
    if (!print_mode) {
    translate([25, 0, 0]) {
        translate([0, 0, base_thickness]) {
            color("Red", 1)
                guide("base_thickness", base_thickness);
            translate([0, 0, base_thickness]) {
                color("PowderBlue", 1)
                    guide("cradle_thickness", cradle_thickness);
                translate([0, 0, cradle_thickness]) {
                    color("Goldenrod", 1)
                        guide("spur_gear_width", spur_gear_width);
                    translate([0, 0, spur_gear_width]) {
                        color("PaleGreen", 1)
                            guide("mid_bar_length", mid_bar_length);
                        translate([0, 0, mid_bar_length]) {
                            color("Tomato", 1)
                                guide("gear_box_base_thickness", gear_box_base_thickness);
                            translate([0, 0, gear_box_base_thickness]) {
                                color("DarkTurquoise", 1)
                                    guide("gear_box_shaft_height", gear_box_shaft_height);
                                translate([0, 0, gear_box_shaft_height]) {
                                    color("LemonChiffon", 1)
                                        guide("gear_box_bottom_gear_clamp_height", gear_box_bottom_gear_clamp_height);
                                    translate([0, 0, gear_box_bottom_gear_clamp_height]) {
                                        color("Thistle", 1)
                                            guide("height_f_gear", height_f_gear);
                                        translate([0, 0, height_f_gear]) {
                                            color("PaleGreen", 1)
                                                guide("pinion_distance", pinion_distance);
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
        color("Cyan", 1)
            guide("base_thickness", base_thickness);
    }
    }



    r = gear_base_radius(modul, gear_teeth);

    crank_handle();


        
        /*
    translate([0, 0, rod_length+27.4]) {
        rotate([0, 0, 90]) {
            render() difference() {
                bevel_gear_pair(modul=modul, gear_teeth=gear_teeth, pinion_teeth=11, axis_angle=90, tooth_width=5, gear_bore=4, pinion_bore=4, pressure_angle = 20, helix_angle=20, together_built=false);
                for (i=[0:4]) {
                    rotate([0, 0, 72*i])
                        translate([7.1, 0, 0])
                            cylinder(r=1.9, h=3);
                }
            }
        }
    }
    */

    translate([0, 0, base_thickness*2+cradle_thickness+spur_gear_width+mid_bar_length+0.2])
        gear_box(modul=1, gear_teeth=30, gear_bore=spur_gear_bore, floor_thickness=gear_box_floor_thickness);

    wall_inner_radius=crank_wall_inner_radius(modul, gear_teeth, crank_handle_washer_height, blade_clearance);
    if (print_mode) {
        translate([40, 20, 0])
            knobbed_bevel_gear(modul, gear_teeth, delta_gear, tooth_width, spur_gear_bore, pressure_angle, helix_angle);
        *translate([0, 40, 0])
            bevel_gear_bottom_clamp(spur_gear_bore, gear_box_bottom_gear_clamp_height);

        *translate([-50, 30, 0])
            bevel_gear_floor(wall_inner_radius, spur_gear_bore, gear_box_wall_thickness, floor_thickness=gear_box_floor_thickness);

        #let (shaft_radius=spur_gear_bore/2) {
            translate([15, -20, 0])
                bevel_gear_washer(spur_gear_bore);
                //washer(shaft_radius+1, shaft_radius, gear_shaft_height+bottom_gear_clamp_height);
        }

        translate([15, -40, 0])
            washer(crank_handle_radius+2, crank_handle_radius, crank_handle_height*2);

        translate([0, -30, 0])
            cylinder(r=1, h=spur_gear_bore+3, $fn=100);

        translate([-30, -5, 0])
            bevel_pinion();

        for (i=[0:2]) {
            translate([i*-10-50, -30, stopper_dim/2])
                drive_shaft_stopper(stopper_dim-0.2, spur_gear_bore*2);
        }
    }

    // Main drive shaft.
    let(
        shaft_length=base_thickness+cradle_thickness+spur_gear_width+mid_bar_length+gear_box_base_thickness+gear_box_shaft_height+gear_box_bottom_gear_clamp_height+
            bevel_gear_cylinder_height(modul, gear_teeth, delta_gear, tooth_width)+top_gear_clamp_height+0
    ) {
    place(
        design_xlate=[0, 0, base_thickness],
        print_xlate=[40, -60, shaft_length],
        print_rotate=[0, 180, 0]
    )
        drive_shaft(shaft_length=shaft_length);
    }

    place(
        design_xlate=[0, 0, base_thickness*2+cradle_thickness],
        print_xlate=[40, -30, 0]
    )
        spur_gear (modul=1, tooth_number=gear_teeth, width=spur_gear_width, bore=spur_gear_bore, pressure_angle=20, helix_angle=0, optimized=true);

    // a cradle to hold the spur gear
    // this rather silly contraption exists because I was adamant about using standardized
    // set of spur gears -- I did not want to create custom gears just for this component.
    // but you are right, it would have been much much easier if I just created a spur gear
    // that had custom interlocking mechanism against the main shaft.
    place(
        design_xlate=[0, 0, base_thickness*2-0.2],
        print_xlate=[-30, -30, 0]
    )
        spur_gear_cradle();

    gear_base(modul, gear_teeth, bore=spur_gear_bore, male=false);

    thickness=2;
    wall_outer_radius=crank_wall_outer_radius(modul, gear_teeth, crank_handle_washer_height, thickness, blade_clearance);
    crank_zloc=base_thickness*2+cradle_thickness+spur_gear_width+mid_bar_length+gear_box_base_thickness+gear_box_shaft_height+gear_box_bottom_gear_clamp_height+height_f_gear;
    wall_outer_normal=crank_wall_outer_normal(modul, gear_teeth, crank_handle_washer_height, thickness, blade_clearance);

/*
    place(
        design_xlate=[0, -(crank_holder_length+wall_outer_radius-3), crank_zloc],
        design_rotate=[-90, 0, 0],
        print_xlate=[0, -30, crank_holder_length+4.2],
        print_rotate=[0, 180, 0]
    )
        crank_handle_shaft_holder();
        */

    place(print_xlate=[0, -50, 0])
    crank_wall(r=r,
        gear_box_floor_zloc=base_thickness*2+cradle_thickness+spur_gear_width+mid_bar_length,
        crank_zloc=crank_zloc);

    if (print_mode) {
        %translate([0, 0, -1])
        cube([300, 300, 2], center=true);
    }
}

    all();
}

gear_toy();
