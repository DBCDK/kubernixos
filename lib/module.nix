{ config, pkgs, lib, ... }: {

  options.kubernixos = with lib; with lib.types; {

    manifests = mkOption {
      type = attrsOf attrs;
      default = {};
      description = "Attribute set of kubernetes manifests.";
    };

    version = mkOption {
      type = enum [ "1.11.0" "1.11.1" ];
      default = "1.11.1";
      description = "Kubernetes version to apply manifest validation against.";
    };

  };

}
