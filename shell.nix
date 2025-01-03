{
  nixpkgs ? import ./nixpkgs.nix,
  pkgs ? import nixpkgs { },
  kubernixos ? import ./default.nix {
    inherit pkgs;
    version = "shell";
  },
  shellHook ?
    let
      # some treefmt formatters are not supported in pre-commit-hooks we
      # filter them out for now.
      toFilter = [ "yamlfmt" ];
      filterFn = n: _v: (!builtins.elem n toFilter);
      treefmtFormatters = pkgs.lib.mapAttrs (_n: v: { inherit (v) enable; }) (
        pkgs.lib.filterAttrs filterFn (import ./treefmt.nix).programs
      );
      pre-commit-hooks = import (
        builtins.fetchTarball "https://github.com/cachix/pre-commit-hooks.nix/tarball/master"
      );
    in
    pre-commit-hooks.run {
      src = ./.;
      hooks = treefmtFormatters;
    },
}:

pkgs.mkShell {
  inherit shellHook;
  inputsFrom = [ kubernixos ];
}
