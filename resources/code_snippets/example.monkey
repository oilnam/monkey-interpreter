// This is a code snippet that demonstrates the use of some of the new features I added.
// ...one being comments ^^

// Another one is a built-in map function that can be used like this:
let doubler = fn(x) {
    x * 2
};

let out = map(doubler, [1,2,1+2])
puts(out) // [2,4,6]

// We have good-old while loops as well:
let foo = fn(n) {
    let x = 0
    while (x < 10) {
        x = x + 1
    }
    return x
}

puts(foo(0))