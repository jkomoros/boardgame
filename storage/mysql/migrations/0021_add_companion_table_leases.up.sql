create table if not exists `companiontableleases` (
  `GameID` varchar(16) not null primary key,
  `Generation` bigint unsigned not null,
  `DeviceID` varchar(128) not null default '',
  `SecretDigest` varchar(64) not null default '',
  `HolderUserID` varchar(128) not null default '',
  `Expires` bigint not null default 0
) engine=InnoDB charset=utf8mb4;
