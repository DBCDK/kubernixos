{ nixpkgs ? import ./nixpkgs.nix
, pkgs ? import nixpkgs {}
}:

let
  kubernixos = import ./default.nix { inherit pkgs; };
in
  pkgs.mkShell {
    name = "kubernixos-build-env";

    inputsFrom = [
      kubernixos
    ];
  }
