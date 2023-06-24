// Local modifications can be tracked via the git log.
// This file's license is independent from github.com/lestrrat-go/openscad.
// Any modifications made to this file should be considered CC-BY-NC-SA,
// the same as the original version. The intention is to 
// make it easier for the upstream project to incorporate these
// changes, should they decide to do so in the future
// (and also to make it easier for us to incorporate their changes!)

$fn = 50;

// Tip clearance
function tip_clearance(modul) = modul / 6;

function transverse_section_helix_angle(pressure_angle, helix_angle) =
    atan(tan(pressure_angle) / cos(helix_angle));

/* Library for Involute Gears, Screws and Racks

This library contains the following modules
- rack(modul, length, height, width, pressure_angle=20, helix_angle=0)
- mountable_rack(modul, length, height, width, pressure_angle=20, helix_angle=0, fastners, profile, head)
- herringbone_rack(modul, length, height, width, pressure_angle = 20, helix_angle=45)
- mountable_herringbone_rack(modul, length, height, width, pressure_angle=20, helix_angle=45, fastners, profile, head)
- spur_gear(modul, tooth_number, width, bore, pressure_angle=20, helix_angle=0, optimized=true)
- herringbone_gear(modul, tooth_number, width, bore, pressure_angle=20, helix_angle=0, optimized=true)
- rack_and_pinion (modul, rack_length, gear_teeth, rack_height, gear_bore, width, pressure_angle=20, helix_angle=0, together_built=true, optimized=true)
- ring_gear(modul, tooth_number, width, rim_width, pressure_angle=20, helix_angle=0)
- herringbone_ring_gear(modul, tooth_number, width, rim_width, pressure_angle=20, helix_angle=0)
- planetary_gear(modul, sun_teeth, planet_teeth, number_planets, width, rim_width, bore, pressure_angle=20, helix_angle=0, together_built=true, optimized=true)
- bevel_gear(modul, tooth_number,  partial_cone_angle, tooth_width, bore, pressure_angle=20, helix_angle=0)
- bevel_herringbone_gear(modul, tooth_number, partial_cone_angle, tooth_width, bore, pressure_angle=20, helix_angle=0)
- bevel_gear_pair(modul, gear_teeth, pinion_teeth, axis_angle=90, tooth_width, bore, pressure_angle = 20, helix_angle=0, together_built=true)
- bevel_herringbone_gear_pair(modul, gear_teeth, pinion_teeth, axis_angle=90, tooth_width, bore, pressure_angle = 20, helix_angle=0, together_built=true)
- worm(modul, thread_starts, length, bore, pressure_angle=20, lead_angle=10, together_built=true)
- worm_gear(modul, tooth_number, thread_starts, width, length, worm_bore, gear_bore, pressure_angle=20, lead_angle=0, optimized=true, together_built=true)

Examples of each module are commented out at the end of this file

Author:     Dr Jörg Janssen
Contributions By:   Keith Emery, Chris Spencer
Last Verified On:      1. June 2018
Version:    2.2
License:     Creative Commons - Attribution, Non Commercial, Share Alike

Permitted modules according to DIN 780:
0.05 0.06 0.08 0.10 0.12 0.16
0.20 0.25 0.3  0.4  0.5  0.6
0.7  0.8  0.9  1    1.25 1.5
2    2.5  3    4    5    6
8    10   12   16   20   25
32   40   50   60

*/


// General Variables
// PI = 3.14159;  // Uncomment if for some reason your OpenSCAD version doesn't have PI defined
rad = 57.29578;
clearance = 0.05;   // clearance between teeth

/*  Converts Radians to Degrees */
function grad(pressure_angle) = pressure_angle*rad;

/*  Converts Degrees to Radians */
function radian(pressure_angle) = pressure_angle/rad;

/*  Converts 2D Polar Coordinates to Cartesian
    Format: radius, phi; phi = Angle to x-Axis on xy-Plane */
function polar_to_cartesian(polvect) = [
    polvect[0]*cos(polvect[1]),
    polvect[0]*sin(polvect[1])
];

/*  Circle Involutes-Function:
    Returns the Polar Coordinates of an Involute Circle
    r = Radius of the Base Circle
    rho = Rolling-angle in Degrees */
function ev(r,rho) = [
    r/cos(rho),
    grad(tan(rho)-radian(rho))
];

/*  Sphere-Involutes-Function
    Returns the Azimuth Angle of an Involute Sphere
    theta0 = Polar Angle of the Cone, where the Cutting Edge of the Large Sphere unrolls the Involute
    theta = Polar Angle for which the Azimuth Angle of the Involute is to be calculated */
function sphere_ev(theta0,theta) = 1/sin(theta0)*acos(cos(theta)/cos(theta0))-acos(tan(theta0)/tan(theta));

/*  Converts Spherical Coordinates to Cartesian
    Format: radius, theta, phi; theta = Angle to z-Axis, phi = Angle to x-Axis on xy-Plane */
function sphere_to_cartesian(vect) = [
    vect[0]*sin(vect[1])*cos(vect[2]),
    vect[0]*sin(vect[1])*sin(vect[2]),
    vect[0]*cos(vect[1])
];

/*  Check if a Number is even
    = 1, if so
    = 0, if the Number is not even */
function is_even(number) =
    (number == floor(number/2)*2) ? 1 : 0;

/*  greatest common Divisor
    according to Euclidean Algorithm.
    Sorting: a must be greater than b */
function ggt(a,b) =
    a%b == 0 ? b : ggt(b,a%b);

/*  Polar function with polar angle and two variables */
function spiral(a, r0, phi) =
    a*phi + r0;

/*  Copy and rotate a Body */
module copier(vect, number, distance, winkel){
    for(i = [0:number-1]){
        translate(v=vect*distance*i)
            rotate(a=i*winkel, v = [0,0,1])
                children(0);
    }
}

/*  rack
    modul = Height of the Tooth Tip above the Rolling LIne
    length = Length of the Rack
    height = Height of the Rack to the Pitch Line
    width = Width of a Tooth
    pressure_angle = Pressure Angle, Standard = 20° according to DIN 867. Should not exceed 45°.
    helix_angle = Helix Angle of the Rack Transverse Axis; 0° = Spur Teeth */
module rack(modul, length, height, width, pressure_angle = 20, helix_angle = 0) {

    // Dimension Calculations
    modul=modul*(1-clearance);
    c = modul / 6;                                              // Tip Clearance
    mx = modul/cos(helix_angle);                          // Module Shift by Helix Angle in the X-Direction
    a = 2*mx*tan(pressure_angle)+c*tan(pressure_angle);       // Flank Width
    b = PI*mx/2-2*mx*tan(pressure_angle);                      // Tip Width
    x = width*tan(helix_angle);                          // Topside Shift by Helix Angle in the X-Direction
    nz = ceil((length+abs(2*x))/(PI*mx));                       // Number of Teeth

    translate([-PI*mx*(nz-1)/2-a-b/2,-modul,0]){
        intersection(){                                         // Creates a Prism that fits into a Geometric Box
            copier([1,0,0], nz, pi*mx, 0){
                polyhedron(
                    points=[[0,-c,0], [a,2*modul,0], [a+b,2*modul,0], [2*a+b,-c,0], [PI*mx,-c,0], [PI*mx,modul-height,0], [0,modul-height,0], // Underside
                        [0+x,-c,width], [a+x,2*modul,width], [a+b+x,2*modul,width], [2*a+b+x,-c,width], [PI*mx+x,-c,width], [PI*mx+x,modul-height,width], [0+x,modul-height,width]],   // Topside
                    faces=[[6,5,4,3,2,1,0],                     // Underside
                        [1,8,7,0],
                        [9,8,1,2],
                        [10,9,2,3],
                        [11,10,3,4],
                        [12,11,4,5],
                        [13,12,5,6],
                        [7,13,6,0],
                        [7,8,9,10,11,12,13],                    // Topside
                    ]
                );
            };
            translate([abs(x),-height+modul-0.5,-0.5]){
                cube([length,height+modul+1,width+1]);          // Cuboid which includes the Volume of the Rack
            }
        };
    };
}

/* Mountable-rack; uses module "rack"
    modul = Height of the Tooth Tip above the Rolling LIne
    length = Length of the Rack
    height = Height of the Rack to the Pitch Line
    width = Width of a Tooth
    pressure_angle = Pressure Angle, Standard = 20° according to DIN 867. Should not exceed 45°.
    helix_angle = Helix Angle of the Rack Transverse Axis; 0° = Spur Teeth
    fastners = Total number of fastners.
    profile = Metric standard profile for fastners (ISO machine screws), M4 = 4, M6 = 6 etc.

    head = Style of fastner to accomodate.
    PH = Pan Head, C = Countersunk, RC = Raised Countersunk, CS = Cap Screw, CSS = Countersunk Socket Screw. */
module mountable_rack(modul, length, height, width, pressure_angle, helix_angle, fastners, profile, head) {
    difference(){
    rack(modul, length, height, width, pressure_angle, helix_angle);
    offset = (length/fastners);
    translate([-length/2+(offset/2),0,0])
    for(i = [0:fastners-1]){
                if (head=="PH"){
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=height+modul, d=profile, center=false);
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=profile*0.6+modul*2.25, d=profile*2, center=false);
                    }
                if (head=="CS"){
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=height+modul, d=profile, center=false);
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=profile*1.25+modul*2.25, d=profile*1.5, center=false);
                    }
                if (head=="C"){
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=height+modul, d=profile, center=false);
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=modul*2.25, d=profile*2, center=false);
                    translate([i*offset,-modul*1.25,width/2])
                    rotate([90,0,0])
                    cylinder (h=profile/2, d1=profile*2, d2=profile, center=false);
                    }
                if (head=="RC"){
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=height+modul, d=profile, center=false);
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=modul*2.25+profile/4, d=profile*2, center=false);
                    translate([i*offset,-modul*1.25-profile/4,width/2])
                    rotate([90,0,0])
                    cylinder (h=profile/2, d1=profile*2, d2=profile, center=false);
                    }
                if (head=="CSS"){
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=height+modul, d=profile, center=false);
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=modul*2.25, d=profile*2, center=false);
                    translate([i*offset,-modul*1.25,width/2])
                    rotate([90,0,0])
                    cylinder (h=profile*0.6, d1=profile*2, d2=profile, center=false);
                    }
                }
            }
        }

// sg = spur gear
function sg_pitch_circle_diameter(modul, tooth_number) =
    modul * tooth_number;
function sg_pitch_circle_radius(modul, tooth_number) =
    let (d = sg_pitch_circle_diameter(modul, tooth_number))
        d / 2;
function sg_base_circle_diameter(modul, tooth_number, pressure_angle, helix_angle) =
    let (
        d = sg_pitch_circle_diameter(modul, tooth_number),
        alpha_spur = transverse_section_helix_angle(pressure_angle, helix_angle)
    )
        d * cos(alpha_spur);
function sg_base_circle_radius(modul, tooth_number, pressure_angle, helix_angle) =
    let (db = sg_base_circle_diameter(modul, tooth_number, pressure_angle, helix_angle))
        db / 2;

function sg_tip_circle_diameter(modul, tooth_number) =
    let (d = sg_pitch_circle_diameter(modul, tooth_number))
        (modul <1)? d + modul * 2.2 : d + modul * 2;

function sg_tip_circle_radius(modul, tooth_number) =
    let (da = sg_tip_circle_diameter(modul, tooth_number))
        da / 2;

function sg_tip_clearance(modul, tooth_number) =
    (tooth_number <3)? 0 : tip_clearance(modul);

function sg_root_circle_diameter(modul, tooth_number) =
    let (
        d = sg_pitch_circle_diameter(modul, tooth_number),
        c = sg_tip_clearance(modul, tooth_number)
    )
        d - 2 * (modul + c);

function sg_root_circle_radius(modul, tooth_number) =
    let (df = sg_root_circle_diameter(modul, tooth_number))
        df / 2;

function sg_max_rolling_angle(modul, tooth_number, pressure_angle, helix_angle) =
    let(
        rb = sg_base_circle_radius(modul, tooth_number, pressure_angle, helix_angle),
        ra = sg_tip_circle_radius(modul, tooth_number)
    )
        acos(rb / ra);

function sg_pitch_circle_rolling_angle(modul, tooth_number, pressure_angle, helix_angle) =
    let (
        rb = sg_base_circle_radius(modul, tooth_number, pressure_angle, helix_angle),
        r  = sg_pitch_circle_radius(modul, tooth_number)
    )
        acos(rb / r);

function sg_involute_pitch_circle_angle(modul, tooth_number, pressure_angle, helix_angle) =
    let (rho_r = sg_pitch_circle_rolling_angle(modul, tooth_number, pressure_angle, helix_angle))
        grad(tan(rho_r)-radian(rho_r));

function sg_torsion_angle(modul, tooth_number, helix_angle, width) =
    let (
        r = sg_pitch_circle_radius(modul, tooth_number),
    )
        rad*width/(r*tan(90-helix_angle));


function sg_involute_steps(modul, tooth_number, pressure_angle, helix_angle) =
    let (rho_ra = sg_max_rolling_angle(modul, tooth_number, pressure_angle, helix_angle))
        rho_ra/16;

function sg_weight_reduction_hole_radius(modul, tooth_number, bore) =
    let(rf = sg_root_circle_radius(modul, tooth_number) )
        (2*rf - bore)/8;

function sg_weight_reduction_hole_distance(modul, tooth_number, bore) =
    let (r_hole = sg_weight_reduction_hole_radius(modul, tooth_number, bore))
        bore/2+2*r_hole;

function sg_weight_reduction_hole_count(modul, tooth_number, bore) =
    let(
        r_hole = sg_weight_reduction_hole_radius(modul, tooth_number, bore),
        rm = sg_weight_reduction_hole_distance(modul, tooth_number, bore)
    )
        floor(2*PI*rm/(3*r_hole));

function sg_should_optimize(want_optimization, modul, tooth_number, width, bore) =
    let (
        d = sg_pitch_circle_diameter(modul, tooth_number),
        r = sg_pitch_circle_radius(modul, tooth_number)
    )
        want_optimization && r >= width*1.5 && d > 2*bore;

function sg_pitch_angle(tooth_number) = 
    360/tooth_number;
/*  Spur gear
    modul = Height of the Tooth Tip beyond the Pitch Circle
    tooth_number = Number of Gear Teeth
    width = tooth_width
    bore = Diameter of the Center Hole
    pressure_angle = Pressure Angle, Standard = 20° according to DIN 867. Should not exceed 45°.
    helix_angle = Helix Angle to the Axis of Rotation; 0° = Spur Teeth
    optimized = Create holes for Material-/Weight-Saving or Surface Enhancements where Geometry allows */
module spur_gear(modul, tooth_number, width, bore, pressure_angle = 20, helix_angle = 0, optimized = true) {
    // Dimension Calculations
    r = sg_pitch_circle_radius(modul, tooth_number);                                                      // Pitch Circle Radius
    rb = sg_base_circle_radius(modul, tooth_number, pressure_angle, helix_angle);                                                    // Base Circle Radius
    rf = sg_root_circle_radius(modul, tooth_number);                                                    // Root Radius
    rho_ra = sg_max_rolling_angle(modul, tooth_number, pressure_angle, helix_angle);                                           // Maximum Rolling Angle;
                                                                    // Involute begins on the Base Circle and ends at the Tip Circle
    rho_r = sg_pitch_circle_rolling_angle(modul, tooth_number, pressure_angle, helix_angle);                                             // Rolling Angle at Pitch Circle;
                                                                    // Involute begins on the Base Circle and ends at the Tip Circle

    phi_r = sg_involute_pitch_circle_angle(modul, tooth_number, pressure_angle, helix_angle);                         // Angle to Point of Involute on Pitch Circle
    gamma = sg_torsion_angle(modul, tooth_number, helix_angle, width);               // Torsion Angle for Extrusion
    step = sg_involute_steps(modul, tooth_number, pressure_angle, helix_angle)
    ;                                            // Involute is divided into 16 pieces
    tau = sg_pitch_angle(tooth_number);                                  // Pitch Angle

    r_hole = sg_weight_reduction_hole_radius(modul, tooth_number, bore); // Radius of Holes for Material-/Weight-Saving

    rm = sg_weight_reduction_hole_distance(modul, tooth_number, bore);   // Distance of the Axes of the Holes from the Main Axis
    z_hole = sg_weight_reduction_hole_count(modul, tooth_number, bore); // Number of Holes for Material-/Weight-Saving

    optimized = sg_should_optimize(optimized, modul, tooth_number, width, bore);

    // Drawing
    union(){
        rotate([0,0,-phi_r-90*(1-clearance)/tooth_number]){                     // Center Tooth on X-Axis;
                                                                        // Makes Alignment with other Gears easier

            linear_extrude(height = width, convexity = 10, twist = gamma){
                difference(){
                    union(){
                        tooth_width = (180*(1-clearance))/tooth_number+2*phi_r;
                        circle(rf);                                     // Root Circle
                        for (rot = [0:tau:360]){
                            rotate (rot){                               // Copy and Rotate "Number of Teeth"
                                polygon(concat(                         // Tooth
                                    [[0,0]],                            // Tooth Segment starts and ends at Origin
                                    [for (rho = [0:step:rho_ra])     // From zero Degrees (Base Circle)
                                                                        // To Maximum Involute Angle (Tip Circle)
                                        polar_to_cartesian(ev(rb,rho))],       // First Involute Flank

                                    [polar_to_cartesian(ev(rb,rho_ra))],       // Point of Involute on Tip Circle

                                    [for (rho = [rho_ra:-step:0])    // of Maximum Involute Angle (Tip Circle)
                                                                        // to zero Degrees (Base Circle)
                                        polar_to_cartesian([ev(rb,rho)[0], tooth_width-ev(rb,rho)[1]])]
                                                                        // Second Involute Flank
                                                                        // (180*(1-clearance)) instead of 180 Degrees,
                                                                        // to allow clearance of the Flanks
                                    )
                                );
                            }
                        }
                    }
                    circle(r = rm+r_hole*1.49);                         // "bore"
                }
            }
        }
        // with Material Savings
        if (optimized) {
            linear_extrude(height = width, convexity = 10){
                difference(){
                        circle(r = (bore+r_hole)/2);
                        circle(r = bore/2);                          // bore
                    }
                }
            linear_extrude(height = (width-r_hole/2 < width*2/3) ? width*2/3 : width-r_hole/2, convexity = 10){
                difference(){
                    circle(r=rm+r_hole*1.51);
                    union(){
                        circle(r=(bore+r_hole)/2);
                        for (i = [0:1:z_hole]){
                            translate(sphere_to_cartesian([rm,90,i*360/z_hole]))
                                circle(r = r_hole);
                        }
                    }
                }
            }
        }
        // without Material Savings
        else {
            linear_extrude(height = width, convexity = 10){
                difference(){
                    circle(r = rm+r_hole*1.51);
                    circle(r = bore/2);
                }
            }
        }
    }
}

/* Herringbone_rack; uses the module "rack"
    modul = Height of the Tooth Tip above the Rolling LIne
    length = Length of the Rack
    height = Height of the Rack to the Pitch Line
    width = Width of a Tooth
    pressure_angle = Pressure Angle, Standard = 20° according to DIN 867. Should not exceed 45°.
    helix_angle = Helix Angle of the Rack Transverse Axis; 0° = Spur Teeth */
module herringbone_rack(modul, length, height, width, pressure_angle = 20, helix_angle) {
 width = width/2;
 translate([0,0,width]){
        union(){
            rack(modul, length, height, width, pressure_angle, helix_angle);      // bottom Half
            mirror([0,0,1]){
                rack(modul, length, height, width, pressure_angle, helix_angle);  // top Half
            }
        }
    }
}

/* Mountable_herringbone_rack; uses module "herringbone_rack"
    modul = Height of the Tooth Tip above the Rolling LIne
    length = Length of the Rack
    height = Height of the Rack to the Pitch Line
    width = Width of a Tooth
    pressure_angle = Pressure Angle, Standard = 20° according to DIN 867. Should not exceed 45°.
    helix_angle = Helix Angle of the Rack Transverse Axis; 0° = Spur Teeth
    fastners = Total number of fastners.
    profile = Metric standard profile for fastners (ISO machine screws), M4 = 4, M6 = 6 etc.

    head = Style of fastner to accomodate.
    PH = Pan Head, C = Countersunk, RC = Raised Countersunk, CS = Cap Screw, CSS = Countersunk Socket Screw. */
module mountable_herringbone_rack(modul, length, height, width, pressure_angle, helix_angle, fastners, profile, head) {
    difference(){
    herringbone_rack(modul, length, height, width, pressure_angle, helix_angle);
    offset = (length/fastners);
    translate([-length/2+(offset/2),0,0])
    for(i = [0:fastners-1]){
                if (head=="PH"){
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=height+modul, d=profile, center=false);
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=profile*0.6+modul*2.25, d=profile*2, center=false);
                    }
                if (head=="CS"){
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=height+modul, d=profile, center=false);
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=profile*1.25+modul*2.25, d=profile*1.5, center=false);
                    }
                if (head=="C"){
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=height+modul, d=profile, center=false);
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=modul*2.25, d=profile*2, center=false);
                    translate([i*offset,-modul*1.25,width/2])
                    rotate([90,0,0])
                    cylinder (h=profile/2, d1=profile*2, d2=profile, center=false);
                    }
                if (head=="RC"){
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=height+modul, d=profile, center=false);
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=modul*2.25+profile/4, d=profile*2, center=false);
                    translate([i*offset,-modul*1.25-profile/4,width/2])
                    rotate([90,0,0])
                    cylinder (h=profile/2, d1=profile*2, d2=profile, center=false);
                    }
                if (head=="CSS"){
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=height+modul, d=profile, center=false);
                    translate([i*offset,modul,width/2])
                    rotate([90,0,0])
                    cylinder(h=modul*2.25, d=profile*2, center=false);
                    translate([i*offset,-modul*1.25,width/2])
                    rotate([90,0,0])
                    cylinder (h=profile*0.6, d1=profile*2, d2=profile, center=false);
                    }
                }
            }
        }

/* Herringbone_gear; uses the module "spur_gear"
    modul = Height of the Tooth Tip beyond the Pitch Circle
    tooth_number = Number of Gear Teeth
    width = tooth_width
    bore = Diameter of the Center Hole
    pressure_angle = Pressure Angle, Standard = 20° according to DIN 867. Should not exceed 45°.
    helix_angle = Helix Angle to the Axis of Rotation, Standard = 0° (Spur Teeth)
    optimized = Holes for Material-/Weight-Saving */
module herringbone_gear(modul, tooth_number, width, bore, pressure_angle = 20, helix_angle=0, optimized=true){

    width = width/2;
    d = modul * tooth_number;                                           // Pitch Circle Diameter
    r = d / 2;                                                      // Pitch Circle Radius
    c =  (tooth_number <3)? 0 : modul/6;                                // Tip Clearance

    df = d - 2 * (modul + c);                                       // Root Circle Diameter
    rf = df / 2;                                                    // Root Radius

    r_hole = (2*rf - bore)/8;                                    // Radius of Holes for Material-/Weight-Saving
    rm = bore/2+2*r_hole;                                        // Distance of the Axes of the Holes from the Main Axis
    z_hole = floor(2*PI*rm/(3*r_hole));                             // Number of Holes for Material-/Weight-Saving

    optimized = (optimized && r >= width*3 && d > 2*bore);      // is Optimization useful?

    translate([0,0,width]){
        union(){
            spur_gear(modul, tooth_number, width, 2*(rm+r_hole*1.49), pressure_angle, helix_angle, false);      // bottom Half
            mirror([0,0,1]){
                spur_gear(modul, tooth_number, width, 2*(rm+r_hole*1.49), pressure_angle, helix_angle, false);  // top Half
            }
        }
    }
    // with Material Savings
    if (optimized) {
        linear_extrude(height = width*2){
            difference(){
                    circle(r = (bore+r_hole)/2);
                    circle(r = bore/2);                          // bore
                }
            }
        linear_extrude(height = (2*width-r_hole/2 < 1.33*width) ? 1.33*width : 2*width-r_hole/2){ //width*4/3
            difference(){
                circle(r=rm+r_hole*1.51);
                union(){
                    circle(r=(bore+r_hole)/2);
                    for (i = [0:1:z_hole]){
                        translate(sphere_to_cartesian([rm,90,i*360/z_hole]))
                            circle(r = r_hole);
                    }
                }
            }
        }
    }
    // without Material Savings
    else {
        linear_extrude(height = width*2){
            difference(){
                circle(r = rm+r_hole*1.51);
                circle(r = bore/2);
            }
        }
    }
}

/*  Rack and Pinion
    modul = Height of the Tooth Tip beyond the Pitch Circle
    rack_length = Length of the Rack
    gear_teeth = Number of Gear Teeth
    rack_height = Height of the Rack to the Pitch Line
    gear_bore = Diameter of the Center Hole of the Spur Gear
    width = Width of a Tooth
    pressure_angle = Pressure Angle, Standard = 20° according to DIN 867. Should not exceed 45°.
    helix_angle = Helix Angle to the Axis of Rotation, Standard = 0° (Spur Teeth) */
module rack_and_pinion (modul, rack_length, gear_teeth, rack_height, gear_bore, width, pressure_angle=20, helix_angle=0, together_built=true, optimized=true) {

    distance = together_built? modul*gear_teeth/2 : modul*gear_teeth;

    rack(modul, rack_length, rack_height, width, pressure_angle, -helix_angle);
    translate([0,distance,0])
        rotate(a=360/gear_teeth)
            spur_gear (modul, gear_teeth, width, gear_bore, pressure_angle, helix_angle, optimized);
}

/*  Ring gear
    modul = Height of the Tooth Tip beyond the Pitch Circle
    tooth_number = Number of Gear Teeth
    width = tooth_width
    rim_width = Width of the Rim from the Root Circle
    bore = Diameter of the Center Hole
    pressure_angle = Pressure Angle, Standard = 20° according to DIN 867. Should not exceed 45°.
    helix_angle = Helix Angle to the Axis of Rotation, Standard = 0° (Spur Teeth) */
module ring_gear(modul, tooth_number, width, rim_width, pressure_angle = 20, helix_angle = 0) {

    // Dimension Calculations
    ha = (tooth_number >= 20) ? 0.02 * atan((tooth_number/15)/pi) : 0.6;    // Shortening Factor of Tooth Head Height
    d = modul * tooth_number;                                           // Pitch Circle Diameter
    r = d / 2;                                                      // Pitch Circle Radius
    alpha_spur = atan(tan(pressure_angle)/cos(helix_angle));// Helix Angle in Transverse Section
    db = d * cos(alpha_spur);                                      // Base Circle Diameter
    rb = db / 2;                                                    // Base Circle Radius
    c = modul / 6;                                                  // Tip Clearance
    da = (modul <1)? d + (modul+c) * 2.2 : d + (modul+c) * 2;       // Tip Diameter
    ra = da / 2;                                                    // Tip Circle Radius
    df = d - 2 * modul * ha;                                        // Root Circle Diameter
    rf = df / 2;                                                    // Root Radius
    rho_ra = acos(rb/ra);                                           // Maximum Involute Angle;
                                                                    // Involute begins on the Base Circle and ends at the Tip Circle
    rho_r = acos(rb/r);                                             // Involute Angle at Pitch Circle;
                                                                    // Involute begins on the Base Circle and ends at the Tip Circle
    phi_r = grad(tan(rho_r)-radian(rho_r));                         // Angle to Point of Involute on Pitch Circle
    gamma = rad*width/(r*tan(90-helix_angle));               // Torsion Angle for Extrusion
    step = rho_ra/16;                                            // Involute is divided into 16 pieces
    tau = 360/tooth_number;                                             // Pitch Angle

    // Drawing
    rotate([0,0,-phi_r-90*(1+clearance)/tooth_number])                      // Center Tooth on X-Axis;
                                                                    // Makes Alignment with other Gears easier
    linear_extrude(height = width, twist = gamma){
        difference(){
            circle(r = ra + rim_width);                            // Outer Circle
            union(){
                tooth_width = (180*(1+clearance))/tooth_number+2*phi_r;
                circle(rf);                                         // Root Circle
                for (rot = [0:tau:360]){
                    rotate (rot) {                                  // Copy and Rotate "Number of Teeth"
                        polygon( concat(
                            [[0,0]],
                            [for (rho = [0:step:rho_ra])         // From zero Degrees (Base Circle)
                                                                    // to Maximum Involute Angle (Tip Circle)
                                polar_to_cartesian(ev(rb,rho))],
                            [polar_to_cartesian(ev(rb,rho_ra))],
                            [for (rho = [rho_ra:-step:0])        // von Maximum Involute Angle (Kopfkreis)
                                                                    // to zero Degrees (Base Circle)
                                polar_to_cartesian([ev(rb,rho)[0], tooth_width-ev(rb,rho)[1]])]
                                                                    // (180*(1+clearance)) statt 180,
                                                                    // to allow clearance of the Flanks
                            )
                        );
                    }
                }
            }
        }
    }

    echo("Ring Gear Outer Diamater = ", 2*(ra + rim_width));

}

/*  Herringbone Ring Gear; uses the Module "ring_gear"
    modul = Height of the Tooth Tip over the Partial Cone
    tooth_number = Number of Gear Teeth
    width = tooth_width
    bore = Diameter of the Center Hole
    pressure_angle = Pressure Angle, Standard = 20° according to DIN 867. Should not exceed 45°.
    helix_angle = Helix Angle to the Axis of Rotation, Standard = 0° (Spur Teeth) */
module herringbone_ring_gear(modul, tooth_number, width, rim_width, pressure_angle = 20, helix_angle = 0) {

    width = width / 2;
    translate([0,0,width])
        union(){
        ring_gear(modul, tooth_number, width, rim_width, pressure_angle, helix_angle);       // bottom Half
        mirror([0,0,1])
            ring_gear(modul, tooth_number, width, rim_width, pressure_angle, helix_angle);   // top Half
    }
}

/*  Planetary Gear; uses the Modules "herringbone_gear" and "herringbone_ring_gear"
    modul = Height of the Tooth Tip over the Partial Cone
    sun_teeth = Number of Teeth of the Sun Gear
    planet_teeth = Number of Teeth of a Planet Gear
    number_planets = Number of Planet Gears. If null, the Function will calculate the Minimum Number
    width = tooth_width
    rim_width = Width of the Rim from the Root Circle
    bore = Diameter of the Center Hole
    pressure_angle = Pressure Angle, Standard = 20° according to DIN 867. Should not exceed 45°.
    helix_angle = Helix Angle to the Axis of Rotation, Standard = 0° (Spur Teeth)
    together_built =
    optimized = Create holes for Material-/Weight-Saving or Surface Enhancements where Geometry allows
    together_built = Components assembled for Construction or separated for 3D-Printing */
module planetary_gear(modul, sun_teeth, planet_teeth, number_planets, width, rim_width, bore, pressure_angle=20, helix_angle=0, together_built=true, optimized=true){

    // Dimension Calculations
    d_sun = modul*sun_teeth;                                     // Sun Pitch Circle Diameter
    d_planet = modul*planet_teeth;                                   // Planet Pitch Circle Diameter
    center_distance = modul*(sun_teeth +  planet_teeth) / 2;        // Distance from Sun- or Ring-Gear Axis to Planet Axis
    ring_teeth = sun_teeth + 2*planet_teeth;              // Number of Teeth of the Ring Gear
    d_ring = modul*ring_teeth;                                 // Ring Pitch Circle Diameter

    rotate = is_even(planet_teeth);                                // Does the Sun Gear need to be rotated?

    n_max = floor(180/asin(modul*(planet_teeth)/(modul*(sun_teeth +  planet_teeth))));
                                                                        // Number of Planet Gears: at most as many as possible without overlap

    // Drawing
    rotate([0,0,180/sun_teeth*rotate]){
        herringbone_gear (modul, sun_teeth, width, bore, pressure_angle, -helix_angle, optimized);      // Sun Gear
    }

    if (together_built){
        if(number_planets==0){
            list = [ for (n=[2 : 1 : n_max]) if ((((ring_teeth+sun_teeth)/n)==floor((ring_teeth+sun_teeth)/n))) n];
            number_planets = list[0];                                      // Determine Number of Planet Gears
             center_distance = modul*(sun_teeth + planet_teeth)/2;      // Distance from Sun- / Ring-Gear Axis
            for(n=[0:1:number_planets-1]){
                translate(sphere_to_cartesian([center_distance,90,360/number_planets*n]))
                    rotate([0,0,n*360*d_sun/d_planet])
                        herringbone_gear (modul, planet_teeth, width, bore, pressure_angle, helix_angle, optimized); // Planet Gears
            }
       }
       else{
            center_distance = modul*(sun_teeth + planet_teeth)/2;       // Distance from Sun- / Ring-Gear Axis
            for(n=[0:1:number_planets-1]){
                translate(sphere_to_cartesian([center_distance,90,360/number_planets*n]))
                rotate([0,0,n*360*d_sun/(d_planet)])
                    herringbone_gear (modul, planet_teeth, width, bore, pressure_angle, helix_angle, optimized); // Planet Gears
            }
        }
    }
    else{
        planet_distance = ring_teeth*modul/2+rim_width+d_planet;     // Distance between Planets
        for(i=[-(number_planets-1):2:(number_planets-1)]){
            translate([planet_distance, d_planet*i,0])
                herringbone_gear (modul, planet_teeth, width, bore, pressure_angle, helix_angle, optimized); // Planet Gears
        }
    }

    herringbone_ring_gear (modul, ring_teeth, width, rim_width, pressure_angle, helix_angle); // Ring Gear

}

// bg = bevel gear
function bg_part_cone_diameter(modul, tooth_number) = modul * tooth_number;

function bg_part_cone_radius(modul, tooth_number) = 
    let(
        d_outside = bg_part_cone_diameter(modul, tooth_number)
     )
        d_outside / 2;

function bg_outside_tooth_cone_radius(modul, tooth_number, partial_cone_angle) = 
    let (
        r_outside = bg_part_cone_radius(modul, tooth_number) 
    )
        r_outside / sin(partial_cone_angle);

function bg_inside_tooth_cone_radius(modul, tooth_number, partial_cone_angle, tooth_width) =
    let (
        rg_outside = bg_outside_tooth_cone_radius(modul, tooth_number, partial_cone_angle)
    )
        rg_outside - tooth_width;

function bg_base_cone_angle(pressure_angle, helix_angle, partial_cone_angle) =
    asin(cos(transverse_section_helix_angle(pressure_angle, helix_angle)) * sin(partial_cone_angle));

function bg_outside_tooth_cone_foot_diameter(modul, tooth_number, partial_cone_angle) =
    let (
        d_outside = bg_part_cone_diameter(modul, tooth_number),
        c = tip_clearance(modul)
    )
        d_outside-(modul+c)*2*cos(partial_cone_angle);

function bg_outside_tooth_cone_foot_radius(modul, tooth_number, partial_cone_angle) =
    let (
        df_outside = bg_outside_tooth_cone_foot_diameter(modul, tooth_number, partial_cone_angle)
    )
        df_outside / 2;

function bg_foot_delta(modul, tooth_number, partial_cone_angle) =
    let (
        rf_outside = bg_outside_tooth_cone_foot_radius(modul, tooth_number, partial_cone_angle),
        rg_outside = bg_outside_tooth_cone_radius(modul, tooth_number, partial_cone_angle)
    )
        asin(rf_outside / rg_outside);

function bg_root_cone_radius(modul, tooth_number, partial_cone_angle) =
    let (
        rg_outside = bg_outside_tooth_cone_radius(modul, tooth_number, partial_cone_angle),
        delta_f = bg_foot_delta(modul, tooth_number, partial_cone_angle),
    )
        rg_outside * sin(delta_f);

function bg_root_cone_height(modul, tooth_number, partial_cone_angle) =
    let (
        rg_outside = bg_outside_tooth_cone_radius(modul, tooth_number, partial_cone_angle),
        delta_f = bg_foot_delta(modul, tooth_number, partial_cone_angle),
    )
        rg_outside * cos(delta_f);

function bg_complimentary_cone_height(modul, tooth_number, partial_cone_angle, tooth_width) =
    let (
        rg_outside = bg_outside_tooth_cone_radius(modul, tooth_number, partial_cone_angle),
    )
        (rg_outside - tooth_width) / cos(partial_cone_angle);

function bg_complimentary_cone_foot_radius(modul, tooth_number, partial_cone_angle, tooth_width) =
    let (
        rg_outside = bg_outside_tooth_cone_radius(modul, tooth_number, partial_cone_angle)
    )
        (rg_outside - tooth_width) / sin(partial_cone_angle);

function bg_complimentary_cone_tip_radius(modul, tooth_number, partial_cone_angle, tooth_width) =
    let (
        rk = bg_complimentary_cone_foot_radius(modul, tooth_number, partial_cone_angle, tooth_width),
        height_k = bg_complimentary_cone_height(modul, tooth_number, partial_cone_angle, tooth_width),
        delta_f = bg_foot_delta(modul, tooth_number, partial_cone_angle)
    )
        rk * height_k * tan(delta_f) / (rk + height_k * tan(delta_f));

function bg_complimentary_truncated_cone_height(modul, tooth_number, partial_cone_angle, tooth_width) =
    let (
        rk = bg_complimentary_cone_foot_radius(modul, tooth_number, partial_cone_angle, tooth_width),
        height_k = bg_complimentary_cone_height(modul, tooth_number, partial_cone_angle, tooth_width),
        delta_f = bg_foot_delta(modul, tooth_number, partial_cone_angle)
    ) 
        rk*height_k/(height_k*tan(delta_f)+rk);

function bevel_gear_cylinder_height(modul, tooth_number, partial_cone_angle, tooth_width) =
    // modul=1, tooth_number=11, partial_cone_angle=100, tooth_width=5 should be 4.75925
    let (
        height_f = bg_root_cone_height(modul, tooth_number, partial_cone_angle),
        height_fk = bg_complimentary_truncated_cone_height(modul, tooth_number, partial_cone_angle, tooth_width)
    )
        height_f - height_fk;


/*  Bevel Gear
    modul = Height of the Tooth Tip over the Partial Cone; Specification for the Outside of the Cone
    tooth_number = Number of Gear Teeth
    partial_cone_angle = (Half)angle of the Cone on which the other Ring Gear rolls
    tooth_width = Width of the Teeth from the Outside toward the apex of the Cone
    bore = Diameter of the Center Hole
    pressure_angle = Pressure Angle, Standard = 20° according to DIN 867. Should not exceed 45°.
    helix_angle = Helix Angle, Standard = 0° */
module bevel_gear(modul, tooth_number, partial_cone_angle, tooth_width, bore, pressure_angle = 20, helix_angle=0) {

    // Dimension Calculations
    d_outside = bg_part_cone_diameter(modul, tooth_number);                                    // Part Cone Diameter at the Cone Base,
                                                                    // corresponds to the Chord in a Spherical Section
    r_outside = bg_part_cone_radius(modul, tooth_number);                                        // Part Cone Radius at the Cone Base
    rg_outside = bg_outside_tooth_cone_radius(modul, tooth_number, partial_cone_angle);                      // Large-Cone Radius for Outside-Tooth, corresponds to the Length of the Cone-Flank;
    rg_inside = bg_inside_tooth_cone_radius(modul, tooth_number, partial_cone_angle, tooth_width);                              // Large-Cone Radius for Inside-Tooth
    delta_b = bg_base_cone_angle(pressure_angle, helix_angle, partial_cone_angle);          // Base Cone Angle

    da_outside = (modul <1)? d_outside + (modul * 2.2) * cos(partial_cone_angle): d_outside + modul * 2 * cos(partial_cone_angle);
    ra_outside = da_outside / 2;
    delta_a = asin(ra_outside/rg_outside);
    c = tip_clearance(modul);                                                  // Tip Clearance

    df_outside = bg_outside_tooth_cone_foot_diameter(modul, tooth_number, partial_cone_angle);
    rf_outside = bg_outside_tooth_cone_foot_radius(modul, tooth_number, partial_cone_angle);
    delta_f = bg_foot_delta(modul, tooth_number, partial_cone_angle);
    rkf = bg_root_cone_radius(modul, tooth_number, partial_cone_angle);                                   // Radius of the Cone Foot
    height_f = bg_root_cone_height(modul, tooth_number, partial_cone_angle);                               // Height of the Cone from the Root Cone

    echo("Part Cone Diameter at the Cone Base = ", d_outside);

    // Sizes for Complementary Truncated Cone
    height_k = bg_complimentary_cone_height(modul, tooth_number, partial_cone_angle, tooth_width);          // Height of the Complementary Cone for corrected Tooth Length

    echo("Complementary Cone Height = ", height_k);
    rk = bg_complimentary_cone_foot_radius(modul, tooth_number, partial_cone_angle, tooth_width);               // Foot Radius of the Complementary Cone
    rfk = bg_complimentary_cone_tip_radius(modul, tooth_number, partial_cone_angle, tooth_width);        // Tip Radius of the Cylinders for
                                                                    // Complementary Truncated Cone
    height_fk = bg_complimentary_truncated_cone_height(modul, tooth_number, partial_cone_angle, tooth_width);                // height of the Complementary Truncated Cones

    echo("Bevel Gear Height (1) = ", height_f-height_fk);
    echo("Bevel Gear Height (2) = ", bevel_gear_cylinder_height(modul, tooth_number, partial_cone_angle, tooth_width));

    phi_r = sphere_ev(delta_b, partial_cone_angle);                      // Angle to Point of Involute on Partial Cone

    // Torsion Angle gamma from Helix Angle
    gamma_g = 2*atan(tooth_width*tan(helix_angle)/(2*rg_outside-tooth_width));
    gamma = 2*asin(rg_outside/r_outside*sin(gamma_g/2));

    step = (delta_a - delta_b)/16;
    tau = 360/tooth_number;                                             // Pitch Angle
    start = (delta_b > delta_f) ? delta_b : delta_f;
    mirrpoint = (180*(1-clearance))/tooth_number+2*phi_r;

    // Drawing
    rotate([0,0,phi_r+90*(1-clearance)/tooth_number]){                      // Center Tooth on X-Axis;
                                                                    // Makes Alignment with other Gears easier
        translate([0,0,height_f]) rotate(a=[0,180,0]){
            union(){
                translate([0,0,height_f]) rotate(a=[0,180,0]){                               // Truncated Cone
                    difference(){
                        linear_extrude(height=height_f-height_fk, scale=rfk/rkf) circle(rkf*1.001); // 1 permille Overlap with Tooth Root
                        translate([0,0,-1]){
                            cylinder(h = height_f-height_fk+2, r = bore/2);                // bore
                        }
                    }
                }
                for (rot = [0:tau:360]){
                    rotate (rot) {                                                          // Copy and Rotate "Number of Teeth"
                        union(){
                            if (delta_b > delta_f){
                                // Tooth Root
                                flankpoint_under = 1*mirrpoint;
                                flankpoint_over = sphere_ev(delta_f, start);
                                polyhedron(
                                    points = [
                                        sphere_to_cartesian([rg_outside, start*1.001, flankpoint_under]),    // 1 permille Overlap with Tooth
                                        sphere_to_cartesian([rg_inside, start*1.001, flankpoint_under+gamma]),
                                        sphere_to_cartesian([rg_inside, start*1.001, mirrpoint-flankpoint_under+gamma]),
                                        sphere_to_cartesian([rg_outside, start*1.001, mirrpoint-flankpoint_under]),
                                        sphere_to_cartesian([rg_outside, delta_f, flankpoint_under]),
                                        sphere_to_cartesian([rg_inside, delta_f, flankpoint_under+gamma]),
                                        sphere_to_cartesian([rg_inside, delta_f, mirrpoint-flankpoint_under+gamma]),
                                        sphere_to_cartesian([rg_outside, delta_f, mirrpoint-flankpoint_under])
                                    ],
                                    faces = [[0,1,2],[0,2,3],[0,4,1],[1,4,5],[1,5,2],[2,5,6],[2,6,3],[3,6,7],[0,3,7],[0,7,4],[4,6,5],[4,7,6]],
                                    convexity =1
                                );
                            }
                            // Tooth
                            for (delta = [start:step:delta_a-step]){
                                flankpoint_under = sphere_ev(delta_b, delta);
                                flankpoint_over = sphere_ev(delta_b, delta+step);
                                polyhedron(
                                    points = [
                                        sphere_to_cartesian([rg_outside, delta, flankpoint_under]),
                                        sphere_to_cartesian([rg_inside, delta, flankpoint_under+gamma]),
                                        sphere_to_cartesian([rg_inside, delta, mirrpoint-flankpoint_under+gamma]),
                                        sphere_to_cartesian([rg_outside, delta, mirrpoint-flankpoint_under]),
                                        sphere_to_cartesian([rg_outside, delta+step, flankpoint_over]),
                                        sphere_to_cartesian([rg_inside, delta+step, flankpoint_over+gamma]),
                                        sphere_to_cartesian([rg_inside, delta+step, mirrpoint-flankpoint_over+gamma]),
                                        sphere_to_cartesian([rg_outside, delta+step, mirrpoint-flankpoint_over])
                                    ],
                                    faces = [[0,1,2],[0,2,3],[0,4,1],[1,4,5],[1,5,2],[2,5,6],[2,6,3],[3,6,7],[0,3,7],[0,7,4],[4,6,5],[4,7,6]],
                                    convexity =1
                                );
                            }
                        }
                    }
                }
            }
        }
    }
}

/*  Bevel Herringbone Gear; uses the Module "bevel_gear"
    modul = Height of the Tooth Tip beyond the Pitch Circle
    tooth_number = Number of Gear Teeth
    partial_cone_angle, tooth_width
    bore = Diameter of the Center Hole
    pressure_angle = Pressure Angle, Standard = 20° according to DIN 867. Should not exceed 45°.
    helix_angle = Helix Angle, Standard = 0° */
module bevel_herringbone_gear(modul, tooth_number, partial_cone_angle, tooth_width, bore, pressure_angle = 20, helix_angle=0){

    // Dimension Calculations

    tooth_width = tooth_width / 2;

    d_outside = modul * tooth_number;                                // Part Cone Diameter at the Cone Base,
                                                                // corresponds to the Chord in a Spherical Section
    r_outside = d_outside / 2;                                    // Part Cone Radius at the Cone Base
    rg_outside = r_outside/sin(partial_cone_angle);                  // Large-Cone Radius, corresponds to the Length of the Cone-Flank;
    c = modul / 6;                                              // Tip Clearance
    df_outside = d_outside - (modul +c) * 2 * cos(partial_cone_angle);
    rf_outside = df_outside / 2;
    delta_f = asin(rf_outside/rg_outside);
    height_f = rg_outside*cos(delta_f);                           // Height of the Cone from the Root Cone

    // Torsion Angle gamma from Helix Angle
    gamma_g = 2*atan(tooth_width*tan(helix_angle)/(2*rg_outside-tooth_width));
    gamma = 2*asin(rg_outside/r_outside*sin(gamma_g/2));

    echo("Part Cone Diameter at the Cone Base = ", d_outside);

    // Sizes for Complementary Truncated Cone
    height_k = (rg_outside-tooth_width)/cos(partial_cone_angle);      // Height of the Complementary Cone for corrected Tooth Length
    rk = (rg_outside-tooth_width)/sin(partial_cone_angle);           // Foot Radius of the Complementary Cone
    rfk = rk*height_k*tan(delta_f)/(rk+height_k*tan(delta_f));    // Tip Radius of the Cylinders for
                                                                // Complementary Truncated Cone
    height_fk = rk*height_k/(height_k*tan(delta_f)+rk);            // height of the Complementary Truncated Cones

    modul_inside = modul*(1-tooth_width/rg_outside);

        union(){
        bevel_gear(modul, tooth_number, partial_cone_angle, tooth_width, bore, pressure_angle, helix_angle);        // bottom Half
        translate([0,0,height_f-height_fk])
            rotate(a=-gamma,v=[0,0,1])
                bevel_gear(modul_inside, tooth_number, partial_cone_angle, tooth_width, bore, pressure_angle, -helix_angle); // top Half
    }
}

/*  Spiral Bevel Gear; uses the Module "bevel_gear"
    modul = Height of the Tooth Tip beyond the Pitch Circle
    tooth_number = Number of Gear Teeth
    height = Height of Gear Teeth
    bore = Diameter of the Center Hole
    pressure_angle = Pressure Angle, Standard = 20° according to DIN 867. Should not exceed 45°.
    helix_angle = Helix Angle, Standard = 0° */
module spiral_bevel_gear(modul, tooth_number, partial_cone_angle, tooth_width, bore, pressure_angle = 20, helix_angle=30){

    steps = 16;

    // Dimension Calculations

    b = tooth_width / steps;
    d_outside = modul * tooth_number;                                // Part Cone Diameter at the Cone Base,
                                                                // corresponds to the Chord in a Spherical Section
    r_outside = d_outside / 2;                                    // Part Cone Radius at the Cone Base
    rg_outside = r_outside/sin(partial_cone_angle);                  // Large-Cone Radius, corresponds to the Length of the Cone-Flank;
    rg_center = rg_outside-tooth_width/2;

    echo("Part Cone Diameter at the Cone Base = ", d_outside);

    a=tan(helix_angle)/rg_center;

    union(){
    for(i=[0:1:steps-1]){
        r = rg_outside-i*b;
        helix_angle = a*r;
        modul_r = modul-b*i/rg_outside;
        translate([0,0,b*cos(partial_cone_angle)*i])

            rotate(a=-helix_angle*i,v=[0,0,1])
                bevel_gear(modul_r, tooth_number, partial_cone_angle, b, bore, pressure_angle, helix_angle);   // top Half
        }
    }
}


// These are tools for the construction of bevel gear pairs.
// The module `bevel_gear_pair` uses these to calculate the correct pair of gears.
//
// They are available here so that users who need to customize the gears can utilize
// the information

// bgp = bevel gear pair

// Cone radius of the gear
function bgp_gear_radius(modul, gear_teeth) = modul * gear_teeth / 2;
// Cone angle of the gear
function bgp_gear_cone_angle(axis_angle, gear_teeth, pinion_teeth) =
    atan(sin(axis_angle)/(pinion_teeth/gear_teeth+cos(axis_angle)));
// Cone angle of the pinion
function bgp_pinion_cone_angle(axis_angle, gear_teeth, pinion_teeth) =
    atan(sin(axis_angle)/(gear_teeth/pinion_teeth+cos(axis_angle)));
// Sphere radius
function bgp_sphere_radius(modul, axis_angle, gear_teeth, pinion_teeth) =
    bgp_gear_radius(modul, gear_teeth) / sin(bgp_gear_cone_angle(axis_angle, gear_teeth, pinion_teeth));
// Bevel diameter on the sphere
function bgp_bevel_diameter(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion=false) =
    PI*bgp_sphere_radius(modul, axis_angle, gear_teeth, pinion_teeth)*
        (is_pinion ?
            bgp_pinion_cone_angle(axis_angle, gear_teeth, pinion_teeth) :
            bgp_gear_cone_angle(axis_angle, gear_teeth, pinion_teeth))
        /90-2*(modul+tip_clearance(modul));
// Root cone radius on the sphere
function bgp_root_cone_radius(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion=false) =
    bgp_bevel_diameter(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion)/2;
// Cone tip angle
function bgp_cone_tip_angle(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion=false) =
    bgp_root_cone_radius(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion)/(PI*bgp_sphere_radius(modul, axis_angle, gear_teeth, pinion_teeth))*180;
// Cone foot angle
function bgp_cone_foot_angle(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion=false) =
    bgp_sphere_radius(modul, axis_angle, gear_teeth, pinion_teeth)*sin(bgp_cone_tip_angle(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion));
// Height of the cone
function bgp_cone_height(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion=false) =
    bgp_sphere_radius(modul, axis_angle, gear_teeth, pinion_teeth)*cos(bgp_cone_tip_angle(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion));

/*  Bevel Gear Pair with any axis_angle; uses the Module "bevel_gear"
    modul = Height of the Tooth Tip over the Partial Cone; Specification for the Outside of the Cone
    gear_teeth = Number of Gear Teeth on the Gear
    pinion_teeth = Number of Gear Teeth on the Pinion
    axis_angle = Angle between the Axles of the Gear and Pinion
    tooth_width = Width of the Teeth from the Outside toward the apex of the Cone
    gear_bore = Diameter of the Center Hole of the Gear
    pinion_bore = Diameter of the Center Bore of the Gear
    pressure_angle = Pressure Angle, Standard = 20° according to DIN 867. Should not exceed 45°.
    helix_angle = Helix Angle, Standard = 0°
    together_built = Components assembled for Construction or separated for 3D-Printing */
module bevel_gear_pair(modul, gear_teeth, pinion_teeth, axis_angle=90, tooth_width, gear_bore, pinion_bore, pressure_angle=20, helix_angle=0, together_built=true){
    // Dimension Calculations
    delta_gear = bgp_gear_cone_angle(axis_angle, gear_teeth, pinion_teeth);    // Cone Angle of the Gear
    delta_pinion = bgp_pinion_cone_angle(axis_angle, gear_teeth, pinion_teeth);// Cone Angle of the Pinion
    delta_f_pinion = bgp_cone_tip_angle(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion=true);               // Tip Cone Angle
    rkf_pinion = bgp_cone_foot_angle(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion=true);                    // Radius of the Cone Foot
    height_f_pinion = bgp_cone_height(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion=true);                // Height of the Cone from the Root Cone

    echo("Cone Angle Gear = ", delta_gear);
    echo("Cone Angle Pinion = ", delta_pinion);

    df_gear = bgp_bevel_diameter(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion=false); // Bevel Diameter on the Large Sphere
    rf_gear = bgp_root_cone_radius(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion=false);                              // Root Cone Radius on the Large Sphere
    rkf_gear = bgp_cone_foot_angle(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion=false);                          // Radius of the Cone Foot
    height_f_gear = bgp_cone_height(modul, axis_angle, gear_teeth, pinion_teeth, is_pinion=false);                      // Height of the Cone from the Root Cone

    echo("Gear Height = ", height_f_gear);
    echo("Pinion Height = ", height_f_pinion);

    do_rotate = is_even(pinion_teeth);

    // Drawing
    // Rad
    rotate([0,0,180*(1-clearance)/gear_teeth*do_rotate])
        bevel_gear(modul, gear_teeth, delta_gear, tooth_width, gear_bore, pressure_angle, helix_angle);

    // Ritzel
    if (together_built)
        translate([-height_f_pinion*cos(90-axis_angle),0,height_f_gear-height_f_pinion*sin(90-axis_angle)])
            rotate([0,axis_angle,0])
                bevel_gear(modul, pinion_teeth, delta_pinion, tooth_width, pinion_bore, pressure_angle, -helix_angle);
    else
        translate([rkf_pinion*2+modul+rkf_gear,0,0])
            bevel_gear(modul, pinion_teeth, delta_pinion, tooth_width, pinion_bore, pressure_angle, -helix_angle);
 }

/*  Herringbone Bevel Gear Pair with arbitrary axis_angle; uses the Module "bevel_herringbone_gear"
    modul = Height of the Tooth Tip over the Partial Cone; Specification for the Outside of the Cone
    gear_teeth = Number of Gear Teeth on the Gear
    pinion_teeth = Number of Gear Teeth on the Pinion
    axis_angle = Angle between the Axles of the Gear and Pinion
    tooth_width = Width of the Teeth from the Outside toward the apex of the Cone
    gear_bore = Diameter of the Center Hole of the Gear
    pinion_bore = Diameter of the Center Bore of the Gear
    pressure_angle = Pressure Angle, Standard = 20° according to DIN 867. Should not exceed 45°.
    helix_angle = Helix Angle, Standard = 0°
    together_built = Components assembled for Construction or separated for 3D-Printing */
module bevel_herringbone_gear_pair(modul, gear_teeth, pinion_teeth, axis_angle=90, tooth_width, gear_bore, pinion_bore, pressure_angle = 20, helix_angle=10, together_built=true){

    r_gear = modul*gear_teeth/2;                           // Cone Radius of the Gear
    delta_gear = atan(sin(axis_angle)/(pinion_teeth/gear_teeth+cos(axis_angle)));   // Cone Angle of the Gear
    delta_pinion = atan(sin(axis_angle)/(gear_teeth/pinion_teeth+cos(axis_angle)));// Cone Angle of the Pinion
    rg = r_gear/sin(delta_gear);                              // Radius of the Large Sphere
    c = modul / 6;                                          // Tip Clearance
    df_pinion = PI*rg*delta_pinion/90 - 2 * (modul + c);    // Bevel Diameter on the Large Sphere
    rf_pinion = df_pinion / 2;                              // Root Cone Radius on the Large Sphere
    delta_f_pinion = rf_pinion/(PI*rg) * 180;               // Tip Cone Angle
    rkf_pinion = rg*sin(delta_f_pinion);                    // Radius of the Cone Foot
    height_f_pinion = rg*cos(delta_f_pinion);                // Height of the Cone from the Root Cone

    echo("Cone Angle Gear = ", delta_gear);
    echo("Cone Angle Pinion = ", delta_pinion);

    df_gear = PI*rg*delta_gear/90 - 2 * (modul + c);          // Bevel Diameter on the Large Sphere
    rf_gear = df_gear / 2;                                    // Root Cone Radius on the Large Sphere
    delta_f_gear = rf_gear/(PI*rg) * 180;                     // Tip Cone Angle
    rkf_gear = rg*sin(delta_f_gear);                          // Radius of the Cone Foot
    height_f_gear = rg*cos(delta_f_gear);                      // Height of the Cone from the Root Cone

    echo("Gear Height = ", height_f_gear);
    echo("Pinion Height = ", height_f_pinion);

    rotate = is_even(pinion_teeth);

    // Gear
    rotate([0,0,180*(1-clearance)/gear_teeth*rotate])
        bevel_herringbone_gear(modul, gear_teeth, delta_gear, tooth_width, gear_bore, pressure_angle, helix_angle);

    // Pinion
    if (together_built)
        translate([-height_f_pinion*cos(90-axis_angle),0,height_f_gear-height_f_pinion*sin(90-axis_angle)])
            rotate([0,axis_angle,0])
                bevel_herringbone_gear(modul, pinion_teeth, delta_pinion, tooth_width, pinion_bore, pressure_angle, -helix_angle);
    else
        translate([rkf_pinion*2+modul+rkf_gear,0,0])
            bevel_herringbone_gear(modul, pinion_teeth, delta_pinion, tooth_width, pinion_bore, pressure_angle, -helix_angle);

}

/*
Archimedean screw.
modul = Height of the Screw Head over the Part Cylinder
thread_starts = Number of Starts (Threads) of the Worm
length = Length of the Worm
bore = Diameter of the Center Hole
pressure_angle = Pressure Angle, Standard = 20° according to DIN 867. Should not exceed 45°.
lead_angle = Lead Angle of the Worm, corresponds to 90° minus Helix Angle. Positive Lead Angle = clockwise.
together_built = Components assembled for Construction or separated for 3D-Printing */
module worm(modul, thread_starts, length, bore, pressure_angle=20, lead_angle, together_built=true){

    // Dimension Calculations
    c = modul / 6;                                              // Tip Clearance
    r = modul*thread_starts/(2*sin(lead_angle));                // Part-Cylinder Radius
    rf = r - modul - c;                                         // Root-Cylinder Radius
    a = modul*thread_starts/(90*tan(pressure_angle));               // Spiralparameter
    tau_max = 180/thread_starts*tan(pressure_angle);                // Angle from Foot to Tip in the Normal Plane
    gamma = -rad*length/((rf+modul+c)*tan(lead_angle));    // Torsion Angle for Extrusion

    step = tau_max/16;

    // Drawing: Extrude with a Twist a Surface enclosed by two Archimedean Spirals
    if (together_built) {
        rotate([0,0,tau_max]){
            linear_extrude(height = length, center = false, convexity = 10, twist = gamma){
                difference(){
                    union(){
                        for(i=[0:1:thread_starts-1]){
                            polygon(
                                concat(
                                    [[0,0]],

                                    // rising Tooth Flank
                                    [for (tau = [0:step:tau_max])
                                        polar_to_cartesian([spiral(a, rf, tau), tau+i*(360/thread_starts)])],

                                    // Tooth Tip
                                    [for (tau = [tau_max:step:180/thread_starts])
                                        polar_to_cartesian([spiral(a, rf, tau_max), tau+i*(360/thread_starts)])],

                                    // descending Tooth Flank
                                    [for (tau = [180/thread_starts:step:(180/thread_starts+tau_max)])
                                        polar_to_cartesian([spiral(a, rf, 180/thread_starts+tau_max-tau), tau+i*(360/thread_starts)])]
                                )
                            );
                        }
                        circle(rf);
                    }
                    circle(bore/2); // Mittelbohrung
                }
            }
        }
    }
    else {
        difference(){
            union(){
                translate([1,r*1.5,0]){
                    rotate([90,0,90])
                        worm(modul, thread_starts, length, bore, pressure_angle, lead_angle, together_built=true);
                }
                translate([length+1,-r*1.5,0]){
                    rotate([90,0,-90])
                        worm(modul, thread_starts, length, bore, pressure_angle, lead_angle, together_built=true);
                    }
                }
            translate([length/2+1,0,-(r+modul+1)/2]){
                    cube([length+2,3*r+2*(r+modul+1),r+modul+1], center = true);
                }
        }
    }
}

/*
Calculates a worm wheel set. The worm wheel is an ordinary spur gear without globoidgeometry.
modul = Height of the screw head above the partial cylinder or the tooth head above the pitch circle
tooth_number = Number of wheel teeth
thread_starts = Number of gears (teeth) of the screw
width = tooth_width
length = Length of the Worm
worm_bore = Diameter of the Center Hole of the Worm
gear_bore = Diameter of the Center Hole of the Spur Gear
pressure_angle = Pressure Angle, Standard = 20° according to DIN 867. Should not exceed 45°.
lead_angle = Pitch angle of the worm corresponds to 90 ° bevel angle. Positive slope angle = clockwise.
optimized = Holes for material / weight savings
together_built =  Components assembled for construction or apart for 3D printing */
module worm_gear(modul, tooth_number, thread_starts, width, length, worm_bore, gear_bore, pressure_angle=20, lead_angle, optimized=true, together_built=true, show_spur=1, show_worm=1){

    c = modul / 6;                                              // Tip Clearance
    r_worm = modul*thread_starts/(2*sin(lead_angle));       // Worm Part-Cylinder Radius
    r_gear = modul*tooth_number/2;                                   // Spur Gear Part-Cone Radius
    rf_worm = r_worm - modul - c;                       // Root-Cylinder Radius
    gamma = -90*width*sin(lead_angle)/(PI*r_gear);         // Spur Gear Rotation Angle
    tooth_distance = modul*pi/cos(lead_angle);                // Tooth Spacing in Transverse Section
    x = is_even(thread_starts)? 0.5 : 1;

    if (together_built) {
        if(show_worm)
        translate([r_worm,(ceil(length/(2*tooth_distance))-x)*tooth_distance,0])
            rotate([90,180/thread_starts,0])
                worm(modul, thread_starts, length, worm_bore, pressure_angle, lead_angle, together_built);

        if(show_spur)
        translate([-r_gear,0,-width/2])
            rotate([0,0,gamma])
                spur_gear (modul, tooth_number, width, gear_bore, pressure_angle, -lead_angle, optimized);
    }
    else {
        if(show_worm)
        worm(modul, thread_starts, length, worm_bore, pressure_angle, lead_angle, together_built);

        if(show_spur)
        translate([-2*r_gear,0,0])
            spur_gear (modul, tooth_number, width, gear_bore, pressure_angle, -lead_angle, optimized);
    }
}

//rack(modul=1, length=60, height=5, width=20, pressure_angle=20, helix_angle=0);

//mountable_rack(modul=1, length=60, height=5, width=20, pressure_angle=20, helix_angle=0, profile=3, head="PH",fastners=3);

//herringbone_rack(modul=1, length=60, height=5, width=20, pressure_angle=20, helix_angle=45);

//mountable_herringbone_rack(modul=1, length=60, height=5, width=20, pressure_angle=20, helix_angle=45, profile=3, head="PH",fastners=3);

//spur_gear (modul=1, tooth_number=30, width=5, bore=4, pressure_angle=20, helix_angle=20, optimized=true);

//herringbone_gear (modul=1, tooth_number=30, width=5, bore=4, pressure_angle=20, helix_angle=30, optimized=true);

//rack_and_pinion (modul=1, rack_length=50, gear_teeth=30, rack_height=4, gear_bore=4, width=5, pressure_angle=20, helix_angle=0, together_built=true, optimized=true);

//ring_gear (modul=1, tooth_number=30, width=5, rim_width=3, pressure_angle=20, helix_angle=20);

//herringbone_ring_gear (modul=1, tooth_number=30, width=5, rim_width=3, pressure_angle=20, helix_angle=30);

//planetary_gear(modul=1, sun_teeth=16, planet_teeth=9, number_planets=5, width=5, rim_width=3, bore=4, pressure_angle=20, helix_angle=30, together_built=true, optimized=true);

//planetary_gear(modul=2, sun_teeth=16, planet_teeth=9, number_planets=5, width=5, rim_width=3, bore=4, pressure_angle=20, helix_angle=30, together_built=true, optimized=false);

//bevel_gear(modul=1, tooth_number=30,  partial_cone_angle=45, tooth_width=5, bore=4, pressure_angle=20, helix_angle=20);

//bevel_herringbone_gear(modul=1, tooth_number=30, partial_cone_angle=45, tooth_width=5, bore=4, pressure_angle=20, helix_angle=30);

//bevel_gear_pair(modul=1, gear_teeth=30, pinion_teeth=11, axis_angle=100, tooth_width=5, gear_bore=4, pinion_bore=4, pressure_angle = 20, helix_angle=20, together_built=true);

//bevel_herringbone_gear_pair(modul=1, gear_teeth=30, pinion_teeth=11, axis_angle=100, tooth_width=5, gear_bore=4, pinion_bore=4, pressure_angle = 20, helix_angle=30, together_built=true);

//worm(modul=1, thread_starts=2, length=15, bore=4, pressure_angle=20, lead_angle=10, together_built=true);

//worm_gear(modul=1, tooth_number=30, thread_starts=2, width=8, length=20, worm_bore=4, gear_bore=4, pressure_angle=20, lead_angle=10, optimized=1, together_built=1, show_spur=1, show_worm=1);
