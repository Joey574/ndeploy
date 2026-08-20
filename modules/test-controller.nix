{ pkgs, ndeploy, ... }:

{
  services.xserver.enable = true;
  services.xserver.windowManager.i3.enable = true;

  users.users.controller = {
    isNormalUser = true;
    password = "";
    extraGroups = [ "video" "input" "wheel" ];
  };

  systemd.tmpfiles.rules = [
    "d /home/controller/.cache 0755 controller users -"
    "d /home/controller/.config 0755 controller users -"
  ];

  services.xserver.displayManager.lightdm.enable = true;
  services.displayManager.autoLogin = {
    enable = true;
    user = "controller";
  };
  services.displayManager.defaultSession = "none+i3";

  services.xserver.windowManager.i3.configFile = pkgs.writeText "i3-config" ''
    exec ${ndeploy}/bin/ndeploy
    bindsym Mod1+f floating toggle
    bindsym Mod1+Shift+q kill
  '';

  virtualisation = {
    graphics = true;
    qemu.options = [
      "-vga none"
      "-device virtio-vga-gl"
      "-display gtk,gl=on,zoom-to-fit=off"
      "-device usb-tablet"
    ];

    cores = 4;
    memorySize = 8192;
    resolution = {
      x = 1920;
      y = 1080;
    };
  };

  environment.systemPackages = with pkgs; [ mesa i3status dmenu ];
  system.stateVersion = "26.05";
}
