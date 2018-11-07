# Kubernixos

This is the first attempt to create a simple wrapper around the declarative
parts of kubectl.

## In a nut-shell

Basically the "kubernixos" script is nothing more than:
`echo $BIG_JSON_BLOB | kubectl apply -f - -l reconciler=kubernixos --prune`

Note the label selector and `--prune` in particular.
This is what ensures that all objects in the cluster, which are **not** present
in the JSON blob get deleted (pruned). This is what makes kubernixos declarative.
In that; only what is declared in the configuration is allowed to
exist in the cluster.

## The module

Kubernixos comes with a NixOS module (`/lib/module.nix`). `kubernixos.manifests`
can hold a list of kubernetes manifests (as attrsets),
which will get serialized into JSON.

*Example:*

```nix
{
  kubernixos.manifests = [{
    kind = "Namespace";
    apiVersion = "v1";
    metadata = {
      name = "development";
      labels = {
        name = "development";
      };
    };
  }];
}
```

## Including Kubernixos in your NixOS config

The kubernixos script accepts two variables to be set at runtime:

1. PACKAGES
2. MODULES

If kubernixos is built using the derivation defined in `/default.nix`,
`PACKAGES` will be set default set to whichever nixpkgs-set is used to build the
derivation. It can however still be overwritten at runtime.

As for `MODULES`, it is important to be able to set modules at runtime, since
this can be used to apply altering options for different environment.

*shell.nix example implementation for morph-managed NixOS config repo:*

```nix
{ pkgs ? import ../../common/nixpkgs.nix { version = "18.09"; } }:

with pkgs.lib;
pkgs.stdenv.mkDerivation {
  name = "kubernixos";

  buildInputs =
  let
    upstream =  with builtins; with pkgs;
      callPackage (fetchFromGitHub {
        owner = "DBCDK";
        repo = "kubernixos";
        rev = "f7d930ef5474255f44af5489eed49648537a386b";
        sha256 = "1wansvk8wqz2c8s4mhvnmyqiyz5mdisris3hvs81bq4zkd2smhbi";
      }) {};

    kubernixos = pkgs.writeShellScriptBin "kubernixos" ''
      if [ ! -f "$1" ]; then
        echo "Hostgroup $1 does not exist!" >&2
        exit 1
      fi

      export MODULES="(import $1).network.modules"
      shift

      ${upstream}/bin/kubernixos $@
    '';
  in
    singleton kubernixos;
}
```

The above example wraps the kubernixos script such that it's first arg becomes
a path to a network-file (similar to hostgroup files used by Morph or NixOps).

Relevant NixOS modules are then passed to kubernixos at runtime and the args
given to the wrapped kubernixos are passed directly all the way to kubectl.
Meaning you can do something like this:
`kubernixos /path/to/hostgroup.nix --dry-run=true`,
and `dry-run=true` is passed to kubectl.
