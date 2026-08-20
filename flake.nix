{
  description = "ndeploy fleet manager";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs, ... }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
      lib = nixpkgs.lib;
    in
    {
      packages.${system}.ndeploy = pkgs.buildGoModule {
        pname = "ndeploy";
        version = "0-unstable";
        src = ./.;

        proxyVendor = true;
        vendorHash = "sha256-HubJA3ZghgufUDq+nvfJxE0MejKl9Z24o+TZyo08930=";
        #vendorHash = lib.fakeHash;

        nativeBuildInputs = with pkgs; [
          pkg-config
        ];

        buildInputs = with pkgs; [
          libGL
          libX11
          libxcursor
          libxrandr
          libxinerama
          libxi
          libxxf86vm
          wayland
          wayland-protocols
          libxkbcommon
        ];
      };

      checks.${system}.fleet-test = pkgs.testers.runNixOSTest {
        name = "ndeploy-test";

        nodes = {
          controller = {pkgs, ... }: {
            imports = [ ./modules/test-controller.nix ];
            environment.systemPackages = [ self.packages.${system}.ndeploy ];
            _module.args.ndeploy = self.packages.${system}.ndeploy;
          };

          node1 = { ... }: { imports = [ ./modules/test-node.nix ]; };
          node2 = { ... }: { imports = [ ./modules/test-node.nix ]; };
        };

        testScript = ''
          start_all()
          node1.wait_for_unit("sshd.service")
          node2.wait_for_unit("sshd.service")
          controller.wait_for_unit("multi-user.target")
          controller.wait_for_unit("display-manager.service")
          controller.sleep(3)
        '';
      };
    };
}
