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

### Submodules

Terralist automatically detects and generates documentation for Terraform submodules within your modules. Submodules are reusable components that follow the [Terraform module structure conventions](https://developer.hashicorp.com/terraform/language/modules/develop/structure).

#### Submodule Discovery

When you upload a module, Terralist automatically scans for submodules in these conventional directories:

- `modules/` - The recommended directory name per Terraform conventions
- `submodules/` - An alternative directory name sometimes used in projects

For example, if your module has this structure:

```
terraform-aws-vpc/
├── main.tf
├── README.md
└── modules/
    ├── vpc-endpoints/
    │   ├── main.tf
    │   └── README.md
    └── flow-log/
        ├── main.tf
        └── README.md
```

Terralist will automatically detect and document both the `modules/vpc-endpoints` and `modules/flow-log` submodules.

#### Nested Submodules

Terralist supports nested submodule structures. For example:

```
modules/
├── networking/
│   ├── vpc/
│   │   └── main.tf
│   └── subnet/
│       └── main.tf
```

Both `modules/networking/vpc` and `modules/networking/subnet` will be detected as separate submodules.

#### Submodule Documentation Generation

For each discovered submodule, Terralist:

1. Checks for a `README.md` file in the submodule directory - if found, uses it as documentation
2. If no README exists, generates documentation automatically from the Terraform files using [terraform-config-inspect](https://github.com/hashicorp/terraform-config-inspect)
3. Stores the documentation with a collision-resistant naming scheme: `{version}__{submodule_path_with_double_underscores}.md`

For example:
- `modules/vpc-endpoints` → `1.0.0__modules__vpc-endpoints.md`
- `modules/networking/vpc` → `1.0.0__modules__networking__vpc.md`

This naming scheme prevents collisions between submodules with similar names (e.g., `modules/net/vpc` vs `modules/net_vpc`).

#### Accessing Submodule Documentation

Submodule documentation can be accessed via API endpoint:

```
GET /v1/api/modules/{name}/{provider}/{version}/submodules/{submodule_path}
```

In the web UI, modules with submodules display a dropdown selector to view documentation for each submodule.

#### Error Handling

If Terralist cannot generate documentation for a submodule (e.g., missing files), it provides a helpful message instead of failing the upload:

```
# Documentation Not Available

No README.md or main.tf file found in this submodule.
```

!!! tip "When organizing your modules, follow the [Terraform module structure conventions](https://developer.hashicorp.com/terraform/language/modules/develop/structure) by placing reusable components in a `modules/` directory. This ensures Terralist can automatically discover and document your submodules."

## Providers

Provider documentation is still in progress. It will be available soon!

## README Rendering Features

Terralist provides rich rendering capabilities for README files with performance-optimized loading:

### Code Syntax Highlighting

Terralist supports syntax highlighting for code blocks using [Shiki](https://shiki.style/), a modern syntax highlighter. Supported languages include:

- **Programming Languages**: JavaScript, TypeScript, Python, Go, Rust, Java, C/C++, and many more
- **Infrastructure as Code**: Terraform (HCL), YAML, JSON, Bash scripts
- **Markup Languages**: HTML, CSS, Markdown

Code blocks are automatically detected and highlighted when present in README files. For optimal performance, syntax highlighting libraries are loaded from CDN only when code blocks are detected in the content.

```terraform
# Example Terraform code block
resource "aws_instance" "example" {
  ami           = "ami-0c55b159cbfafe1d0"
  instance_type = "t2.micro"

  tags = {
    Name = "Example instance"
  }
}
```

### Mermaid Diagrams

Terralist supports diagram creation using [Mermaid](https://mermaid.js.org/), a JavaScript-based diagramming library. Supported diagram types include:

- **Flowcharts**: Graph and flowchart diagrams
- **Sequence Diagrams**: Interaction diagrams
- **State Diagrams**: Finite state machine diagrams
- **Class Diagrams**: Object-oriented design diagrams
- **Entity Relationship Diagrams**: Database schema diagrams

Mermaid diagrams are automatically detected and rendered when present in README files. For optimal performance, the Mermaid library is loaded from CDN only when diagram code blocks are detected in the content.

```mermaid
graph TD;
    A[Start] --> B[Process];
    B --> C[Validate];
    C --> D[Deploy];
    D --> E[End];
```

### Emoji Shortcode Support

README files support emoji shortcodes using the [node-emoji](https://github.com/omnidan/node-emoji) library. You can use GitHub-style emoji shortcodes that are automatically converted to Unicode emojis:

- `:smile:` → 😄
- `:rocket:` → 🚀
- `:warning:` → ⚠️
- `:bulb:` → 💡
