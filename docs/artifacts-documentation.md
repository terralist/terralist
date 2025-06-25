# Artifacts Documentation

Terralist generates and renders documentation for the artifacts it stores. You can either bring your own documentation or let Terralist generate it automatically.

## Modules

Module documentation consists of a single Markdown file. Users can bring their own documentation by bundling a `README.md` file with the module before uploading it to Terralist.

If you'd like Terralist to automatically generate module documentation for you, here's what happens:

1. Terralist unzips your module files locally.
2. The files are processed using [terraform-config-inspect](https://github.com/hashicorp/terraform-config-inspect).
3. The generated Markdown is pushed to the storage provider and served on each request (either rendered in the UI or returned via the API).

!!! warning "If, by any means, Terralist is unable to process the module and generate the documentation for it, the upload will NOT fail. The module archive will still be uploaded without documentation and a warning log will be produced."

Terralist will attempt to find a `README.md` before generating the documentation on its own. To properly detect which `README.md` file is the correct one, it will recursively traverse the directory tree. The first parent node which contains any `*.tf` file will be considered a root module, and, if this directory contains a `README.md` file, the file will be selected and used as documentation.

!!! note "If the archive contains multiple subdirectories, and at least two of them have `*.tf` files (and a `README.md` file), it is undetermined which one will be selected as the 'root module' - depending on how the OS sorts the directories."

## Providers

Provider documentation is still in progress. It will be available soon!
