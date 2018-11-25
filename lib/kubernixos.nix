 { packages ? <nixpkgs>, modules ? [] }:
let
  pkgs = if builtins.isAttrs packages then packages else (import packages {});

  cfg = (import "${toString pkgs.path}/nixos/lib/eval-config.nix" {
    inherit pkgs modules;
  }).config.kubernixos;

  kubernixos = with builtins; substring 0 32 (hashString "sha256" (toJSON cfg.manifests));

  merge = name: item:
    pkgs.lib.recursiveUpdate item {
      metadata = {
        labels = {
          inherit kubernixos;
        };
      };
    };

in
{

  kubernixos = {

    config = (removeAttrs cfg ["assertions" "manifests"]) // { checksum = kubernixos; };

    manifests = with pkgs;
    let
      # Assertion validation borrowed from /modules/system/activation/top-level.nix
      failedAssertions = with pkgs.lib; map (x: x.message) (filter (x: !x.assertion) cfg.assertions);
    in
      if failedAssertions != []
      then throw "\nFailed assertions:\n${concatStringsSep "\n" (map (x: "- ${x}") failedAssertions)}"
      else
      {
        apiVersion = "v1";
        kind = "List";
        items = lib.mapAttrsToList merge cfg.manifests;
      };

  };
}
