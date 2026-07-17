alter table `companiontableleases`
  drop column `TransitionKind`,
  drop column `PreviousDeviceID`,
  drop column `TransferTargetDeviceID`,
  drop column `TransferExpires`,
  drop column `TransferCodeDigest`,
  drop column `TransferTokenDigest`,
  drop column `TransferID`;
