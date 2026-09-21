{ pkgs, ndeploy, ... }:

{
  imports = [
    ./ssh.nix
  ];

  services.xserver.enable = true;
  services.xserver.windowManager.i3.enable = true;

  users.users.controller = {
    isNormalUser = true;
    password = "";
    extraGroups = [
      "video"
      "input"
      "wheel"
    ];
  };

  systemd.tmpfiles.rules = [
    "d /home/controller/.cache 0755 controller users -"
    "d /home/controller/.config 0755 controller users -"
  ];

  services.displayManager.sddm.enable = true;
  services.desktopManager.plasma6.enable = true;
  services.displayManager.defaultSession = "plasma";

  services.displayManager.autoLogin = {
    enable = true;
    user = "controller";
  };

  virtualisation = {
    graphics = true;
    qemu.options = [
      "-vga none"
      "-device virtio-vga-gl"
      "-display gtk,gl=on,zoom-to-fit=off"
      "-device usb-tablet"
    ];

    cores = 6;
    memorySize = 16384;
    resolution = {
      x = 1920;
      y = 1080;
    };
  };

  environment.systemPackages = with pkgs; [
    mesa
    ndeploy
  ];
  system.stateVersion = "26.05";
}
