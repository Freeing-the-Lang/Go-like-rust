trait Drawable {
    fn draw(self);
}

struct Point {
    x: i32;
}

impl Drawable for Point {
    fn draw(self) {
        println!(self.x);
    }
}

fn main() {
    let p = 42;
    println!(p);
}
