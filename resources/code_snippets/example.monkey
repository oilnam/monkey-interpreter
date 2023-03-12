// This is a code snippet that demonstrates the use of some of the new features I added.
// ...one being comments ^^

// Another one is a built-in map function that can be used like this:
let doubler = fn(x) {
    x * 2
};

let out = map(doubler, [1,2,1+2])
puts(out) // [2,4,6]

// We have good-old while loops as well:
let x = 10;
while (x < 100) {
    let x = x + 10 // TODO: support variable reassignment
    puts(x)
}
