# cf-emulator

To run:

  1. Create a directory ```c:\containerizer```
  2. Clone and build Containerizer and place build output in ```c:\containerizer```. [`windows garden`](https://github.com/cloudfoundry/garden-windows)
  3. Clone and build hwc.exe and copy to ```c:\containerizer```. [`hwc`](https://github.com/cloudfoundry-incubator/hwc-buildpack)
  4. Enable disk quotas on C DRIVE.
  5. Ensure windows firewall is enabled.
  6. Run cf-emulator: ```cf-emulator --path c:\Projects\MyApplication```

Currently the emulator uses a hardcoded container handle which results in the same hash (1409DF6E45C1C89E09) for the directory and user. To rerun the application you must do the following:

1. Delete the local user `c_1409DF6E45C1C89E09`.
2. Delete the directory ```c:\containerizer\1409DF6E45C1C89E09```
3. Delete the port reservation via netsh ```netsh http delete urlacl http://*:64055/```

The port is currently always 64055.