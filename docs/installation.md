# Installation

To install Terralist on your machine, select below your operating system and your CPU architecture.

=== "Linux"
    <div style="font-size: 18px; font-weight: bold;">Download</div>

    === "amd64"

        ``` bash
        curl -sL https://github.com/terralist/terralist/releases/latest/download/terralist_linux_amd64.zip -o terralist_linux_amd64.zip
        unzip terralist_linux_amd64.zip
        ```

    === "i386"

        ``` bash
        curl -sL https://github.com/terralist/terralist/releases/latest/download/terralist_linux_386.zip -o terralist_linux_386.zip
        unzip terralist_linux_386.zip
        ```

    === "arm64"

        ``` bash
        curl -sL https://github.com/terralist/terralist/releases/latest/download/terralist_linux_arm64.zip -o terralist_linux_arm64.zip
        unzip terralist_linux_arm64.zip
        ```

    <div style="font-size: 18px; font-weight: bold;">Usage</div>

    ``` bash
    ./terralist --help
    ```

=== "MacOS"
    <div style="font-size: 18px; font-weight: bold;">Download</div>

    === "amd64"

        ``` bash
        curl -sL https://github.com/terralist/terralist/releases/latest/download/terralist_linux_amd64.zip -o terralist_darwin_amd64.zip
        unzip terralist_darwin_amd64.zip
        ```
    === "arm64"

        ``` bash
        curl -sL https://github.com/terralist/terralist/releases/latest/download/terralist_darwin_arm64.zip -o terralist_darwin_arm64.zip
        unzip terralist_linux_arm64.zip
        ```

    <div style="font-size: 18px; font-weight: bold;">Usage</div>

    ``` bash
    ./terralist --help
    ```

=== "Windows"
    <div style="font-size: 18px; font-weight: bold;">Download</div>

    === "amd64"

        ``` powershell
        Invoke-WebRequest "https://github.com/terralist/terralist/releases/latest/download/terralist_windows_amd64.zip" -OutFile "terralist_windows_amd64.zip"
        Expand-Archive terralist_windows_amd64.zip -DestinationPath terralist
        ```
  
    === "i386"

        ``` powershell
        Invoke-WebRequest "https://github.com/terralist/terralist/releases/latest/download/terralist_windows_386.zip" -OutFile "terralist_windows_386.zip"
        Expand-Archive terralist_windows_386.zip -DestinationPath terralist
        ```

    <div style="font-size: 18px; font-weight: bold;">Usage</div>

    ``` powershell
    .\terralist.exe --help
    ```

=== "Docker"
    <div style="font-size: 18px; font-weight: bold;">Download</div>

    ``` bash
    docker pull ghcr.io/terralist/terralist
    ```

    <div style="font-size: 18px; font-weight: bold;">Usage</div>

    ``` bash
    docker run ghcr.io/terralist/terralist --help
    ```


If you're following this documentation as a step-by-step guide, you may proceed with the [Getting Started](./getting-started.md) guide.
