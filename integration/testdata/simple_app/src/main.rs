// Steps to install this program:
// 1. curl https://sh.rustup.rs -sSf | sh
// 2. source $HOME/.cargo/env
// 3. rustup default nightly
// 4. cargo run

#![feature(proc_macro_hygiene, decl_macro)]

#[macro_use] extern crate rocket;

#[get("/")]
fn index() -> &'static str {
    "Hello, world!"
}

fn main() {
    rocket::ignite().mount("/", routes![index]).launch();
}
