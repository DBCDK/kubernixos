{ pkgs ? (import <nixpkgs> {}) }:
# TODO: PIN pkgs
let
  kubernixos = import ./default.nix { inherit pkgs; };
in
  pkgs.mkShell {
    name = "kubernixos-build-env";

    inputsFrom = [
      kubernixos
    ];
  }
