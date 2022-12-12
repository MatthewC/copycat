# Copy Cat - Environment Manager

CopyCat is a `.env` manager I used as an excuse to learn [Go](https://go.dev). You can configure it with your own S3 (or MinIO) instance to have .env files synced. The client is also capable of syncing any files to a specific environment, so that they can later be re-downloaded. 

## Installation
Asuming Go has already been installed, you should be able to run

```
make build
make install
```
to install the binary. There is a known issue with Windows that I haven't been able to figure out with the `Syscall.umask` function, if you are building on Windows I'd recommend commenting out [this line](https://github.com/MatthewC/copycat/blob/57f1e2ffaf36d1b4e6c9a3726af4f0ac22a11d14/commands.go#L44).

## Usage
Use examples liberally, and show the expected output if you can. It's helpful to have inline the smallest example of usage that you can demonstrate, while providing links to more sophisticated examples if they are too long to reasonably include in the README.

## Support
If you encounter any issue with the binary, feel free to open an Issue and I'll take a look at it as soon as I can.

## Contributing
Contributions are always welcome! Feel free to fork and later open a pull request explaining what your changes do.

## License
See [License](LICENSE)

## Project status
There are still things I want to add to this project (most notably, test cases), but as of now, development of this project has slowed down.
