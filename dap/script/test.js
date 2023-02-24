var g_a = 1;
var g_b = "abc";
var g_c = null;
var g_d = false;
var g_e = [1, 2];
var g_f = {a: 'foo', b: 42, c: {}};

function f1() {
    let a = 1;
    let b = "abc";
    let c = null;
    let d = false;
    let e = [1, 2];
    let f = {a: 'foo', b: 42, c: {}};
    a++;
    return 1;
}
f1();