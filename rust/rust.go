package rust

const Dependency = "rust"

// TODO Implement Command sequence
// 1. cd rustup*
// 2. sh rustup-init.sh -y
// 3. source $HOME/.cargo/env
// 4. parse rust version from buildpack.yml
// 5. rustup default <version>
// 6. cargo run

// TODO hash Cargo.lock to determine caching
