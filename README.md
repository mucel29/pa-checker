# PA CHECKER

## Project Structure

* root
  * `bin` - Windows & Linux compiled binaries 
  * `res` - project resources: config files
  * `src` - project source files
    * `checker-modules`
    * `interface-modules`
    * `manager`
    * `utils`
  * `main.go` - project entrypoint
  * `Makefile` - use this to compile the project

## Contributing

* Keep in mind that Andra recommended that the commits be in english.
* The commits must be signed (`git commit -s`)
* The commit messages should have the following structure

```
MODULE: <concise title>

*detailed description* (around 75 characters per line)
```
> Example commit message
> ```
> ref: Added order checks
> 
> Lorem ipsum odor amet, consectetuer adipiscing elit. Neque magna platea
> ornare a maecenas aptent tincidunt. Tellus dolor maecenas congue pharetra
> leo himenaeos dis curabitur. Accumsan venenatis eget ipsum enim montes
> volutpat quisque. Diam finibus leo mattis fames efficitur.
> ```

## Roadmap

* Package names 
* Logging
* More linters?
* Better build script _(Added more checks when compiling on Windows or Linux)_
* Unit tests

## Low priority
* Better output when a pipeline fails
* [Commit hooks?](https://pre-commit.com/)
