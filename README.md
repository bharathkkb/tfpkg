# tfpkg

tfpkg contains a CLI and library tfgen for building TF root configurations by composing Terraform modules. Some samples are available [here](https://github.com/bharathkkb/tfpkg-samples).

To install the CLI:

```sh
go install github.com/bharathkkb/tfpkg@latest
```

To generate a module pkg:

```
tfpkg ${module-registry-source}
tfpkg terraform-google-modules/network/google
```

Some samples are available [here](https://github.com/bharathkkb/tfpkg-samples) utilizing both tfpkg and tfgen.
