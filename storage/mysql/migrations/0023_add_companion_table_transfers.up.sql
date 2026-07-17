alter table `companiontableleases`
  add column `TransferID` varchar(32) not null default '',
  add column `TransferTokenDigest` varchar(64) not null default '',
  add column `TransferCodeDigest` varchar(64) not null default '',
  add column `TransferExpires` bigint not null default 0,
  add column `TransferTargetDeviceID` varchar(32) not null default '',
  add column `PreviousDeviceID` varchar(32) not null default '',
  add column `TransitionKind` varchar(16) not null default '';
