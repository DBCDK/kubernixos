{ config, pkgs, lib, ... }:

let
  cfg = config.kubernixos;
in
{

  options.kubernixos = with lib; with lib.types; {

    manifests = mkOption {
      type = attrsOf attrs;
      default = {};
      description = "Attribute set of kubernetes manifests.";
    };

    schemas = mkOption {
      type = package;
      default = pkgs.kubeval-schema;
      description = "k8s jsonschemas package";
    };

    version = mkOption {
      type = str;
      description = "Kubernetes version to apply manifest validation against.";
    };

    server = mkOption {
      type = str;
      description = ''
        Address to the kubernetes apiserver.
        This is a safety to ensure that applying kubeconfigs and manifest conf match.
      '';
    };

  };
}
