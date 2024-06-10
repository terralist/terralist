SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

SET NAMES utf8mb4;

DROP DATABASE IF EXISTS `terralist`;
CREATE DATABASE `terralist` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;
USE `terralist`;

DROP TABLE IF EXISTS `authorities`;
CREATE TABLE `authorities` (
  `id` varchar(256) NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `name` varchar(256) NOT NULL,
  `policy_url` varchar(256) NOT NULL,
  `owner` varchar(256) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_authorities_name` (`name`),
  KEY `idx_authorities_owner` (`owner`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO `authorities` (`id`, `created_at`, `updated_at`, `name`, `policy_url`, `owner`) VALUES
('04d7980b-9cdd-4cec-bc80-46db639e18b3',	'2024-04-19 13:42:36',	'2024-04-19 13:53:57',	'hashicorp',	'https://www.hashicorp.com/security.html',	''),
('9a50dba6-7ab3-4ee6-8660-8f5901337883',	'2024-04-19 13:42:02',	'2024-04-19 13:42:02',	'terraform-aws-modules',	'',	'');

DROP TABLE IF EXISTS `authority_api_keys`;
CREATE TABLE `authority_api_keys` (
  `id` varchar(256) NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `authority_id` varchar(256) DEFAULT NULL,
  `expiration` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_authorities_api_keys` (`authority_id`),
  CONSTRAINT `fk_authorities_api_keys` FOREIGN KEY (`authority_id`) REFERENCES `authorities` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO `authority_api_keys` (`id`, `created_at`, `updated_at`, `authority_id`, `expiration`) VALUES
('7684e467-4038-4b04-b03b-eec76521c7c2',	'2024-04-19 13:51:18',	'2024-04-19 13:51:18',	'04d7980b-9cdd-4cec-bc80-46db639e18b3',	NULL),
('b2eeb8e8-f318-442c-8237-962cad073496',	'2024-04-19 13:49:25',	'2024-04-19 13:49:25',	'9a50dba6-7ab3-4ee6-8660-8f5901337883',	NULL);

DROP TABLE IF EXISTS `authority_keys`;
CREATE TABLE `authority_keys` (
  `id` varchar(256) NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `authority_id` varchar(256) DEFAULT NULL,
  `key_id` varchar(256) NOT NULL,
  `ascii_armor` longtext,
  `trust_signature` longtext,
  PRIMARY KEY (`id`),
  KEY `fk_authorities_keys` (`authority_id`),
  CONSTRAINT `fk_authorities_keys` FOREIGN KEY (`authority_id`) REFERENCES `authorities` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO `authority_keys` (`id`, `created_at`, `updated_at`, `authority_id`, `key_id`, `ascii_armor`, `trust_signature`) VALUES
('df11cb4a-b8c7-4b32-9802-b70fcdda729f',	'2024-04-19 13:53:57',	'2024-04-19 13:53:57',	'04d7980b-9cdd-4cec-bc80-46db639e18b3',	'34365D9472D7468F',	'-----BEGIN PGP PUBLIC KEY BLOCK-----\n\n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\nZF5q4h4I33PSGDdSvGXn9UMY5Isjpg==\n=7pIB\n-----END PGP PUBLIC KEY BLOCK-----',	'');

DROP TABLE IF EXISTS `module_dependencies`;
CREATE TABLE `module_dependencies` (
  `id` varchar(256) NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `parent_id` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_module_versions_dependencies` (`parent_id`),
  CONSTRAINT `fk_module_submodules_dependencies` FOREIGN KEY (`parent_id`) REFERENCES `module_submodules` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_module_versions_dependencies` FOREIGN KEY (`parent_id`) REFERENCES `module_versions` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


DROP TABLE IF EXISTS `module_providers`;
CREATE TABLE `module_providers` (
  `id` varchar(256) NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `parent_id` varchar(256) DEFAULT NULL,
  `name` varchar(256) NOT NULL,
  `namespace` varchar(256) NOT NULL,
  `source` varchar(256) NOT NULL,
  `version` varchar(256) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_module_submodules_providers` (`parent_id`),
  CONSTRAINT `fk_module_submodules_providers` FOREIGN KEY (`parent_id`) REFERENCES `module_submodules` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_module_versions_providers` FOREIGN KEY (`parent_id`) REFERENCES `module_versions` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


DROP TABLE IF EXISTS `module_submodules`;
CREATE TABLE `module_submodules` (
  `id` varchar(256) NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `version_id` varchar(256) DEFAULT NULL,
  `path` varchar(256) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_module_versions_submodules` (`version_id`),
  CONSTRAINT `fk_module_versions_submodules` FOREIGN KEY (`version_id`) REFERENCES `module_versions` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


DROP TABLE IF EXISTS `module_versions`;
CREATE TABLE `module_versions` (
  `id` varchar(256) NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `module_id` varchar(256) DEFAULT NULL,
  `version` varchar(256) NOT NULL,
  `location` varchar(256) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_modules_versions` (`module_id`),
  CONSTRAINT `fk_modules_versions` FOREIGN KEY (`module_id`) REFERENCES `modules` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO `module_versions` (`id`, `created_at`, `updated_at`, `module_id`, `version`, `location`) VALUES
('017b725a-3200-4c36-8020-0742223429a7',	'2024-04-19 14:08:34',	'2024-04-19 14:08:34',	'1f0e1ead-d4b3-4c4b-9313-9fa4a9cf15fe',	'4.1.2',	'modules/terraform-aws-modules/s3-bucket/aws/4.1.2.zip'),
('0c8e4b8c-f025-4560-9803-7b8c519c279e',	'2024-04-19 13:49:53',	'2024-04-19 13:49:53',	'756e52be-4242-4a2a-9d65-b969f47ff978',	'5.6.0',	'modules/terraform-aws-modules/vpc/aws/5.6.0.zip'),
('17647521-e531-4a83-b690-cb43ac87ba20',	'2024-04-19 13:50:05',	'2024-04-19 13:50:05',	'41cc61e4-e14a-4ed3-a20e-a47160194a53',	'6.5.4',	'modules/terraform-aws-modules/rds/aws/6.5.4.zip'),
('1b8cee2b-5c72-4904-86a9-63a42368c40c',	'2024-04-19 13:50:03',	'2024-04-19 13:50:03',	'51c70b4c-57e1-4795-8f9e-19e7adba3e58',	'7.2.6',	'modules/terraform-aws-modules/lambda/aws/7.2.6.zip'),
('3049393a-3c01-4870-8898-5fc61f6d3d6f',	'2024-04-19 13:49:58',	'2024-04-19 13:49:58',	'7ea71bb3-63e7-4a20-8f5b-6ec67a241859',	'20.8.3',	'modules/terraform-aws-modules/eks/aws/20.8.3.zip'),
('4b272992-e913-4aa2-9164-d8962a3a6489',	'2024-04-19 13:49:56',	'2024-04-19 13:49:56',	'7ea71bb3-63e7-4a20-8f5b-6ec67a241859',	'20.8.1',	'modules/terraform-aws-modules/eks/aws/20.8.1.zip'),
('5bd0c46b-09db-481f-9211-e7db7a883122',	'2024-04-19 13:49:54',	'2024-04-19 13:49:54',	'756e52be-4242-4a2a-9d65-b969f47ff978',	'5.7.0',	'modules/terraform-aws-modules/vpc/aws/5.7.0.zip'),
('602d385d-7ca3-46e1-bbfc-c61b2dca69f0',	'2024-04-19 13:49:55',	'2024-04-19 13:49:55',	'7ea71bb3-63e7-4a20-8f5b-6ec67a241859',	'20.8.0',	'modules/terraform-aws-modules/eks/aws/20.8.0.zip'),
('6ad6ac12-89e6-4ec7-b322-493f85f2f7ad',	'2024-04-19 13:49:57',	'2024-04-19 13:49:57',	'7ea71bb3-63e7-4a20-8f5b-6ec67a241859',	'20.8.2',	'modules/terraform-aws-modules/eks/aws/20.8.2.zip'),
('7145990b-3422-4ac0-9b64-782cd50a8494',	'2024-04-19 13:49:52',	'2024-04-19 13:49:52',	'756e52be-4242-4a2a-9d65-b969f47ff978',	'5.5.3',	'modules/terraform-aws-modules/vpc/aws/5.5.3.zip'),
('85b5ccad-79fd-4413-a61e-4df562fd4084',	'2024-04-19 13:50:01',	'2024-04-19 13:50:01',	'4fa3edc9-4c2b-42fa-bf43-c9c02240f87d',	'5.1.2',	'modules/terraform-aws-modules/security-group/aws/5.1.2.zip'),
('8a37151b-9adb-48f3-bd7e-baad8d8308ad',	'2024-04-19 14:08:33',	'2024-04-19 14:08:33',	'0e0fe8e0-e43f-4e8e-bf4a-893e6d3237b3',	'2.2.1',	'modules/terraform-aws-modules/kms/aws/2.2.1.zip'),
('98f8afdc-7c04-4460-b244-fa520e76a6f2',	'2024-04-19 13:49:54',	'2024-04-19 13:49:54',	'756e52be-4242-4a2a-9d65-b969f47ff978',	'5.7.1',	'modules/terraform-aws-modules/vpc/aws/5.7.1.zip'),
('a11de7e8-eb30-40f4-9fb4-d88582323cc9',	'2024-04-19 13:50:00',	'2024-04-19 13:50:00',	'4fa3edc9-4c2b-42fa-bf43-c9c02240f87d',	'5.1.1',	'modules/terraform-aws-modules/security-group/aws/5.1.1.zip'),
('f070736f-2761-4b4a-a789-0593170a382b',	'2024-04-19 13:50:02',	'2024-04-19 13:50:02',	'4bcefa00-6beb-40b9-8740-d13aa6c8c7c3',	'3.4.0',	'modules/terraform-aws-modules/cloudfront/aws/3.4.0.zip');

DROP TABLE IF EXISTS `modules`;
CREATE TABLE `modules` (
  `id` varchar(256) NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `authority_id` varchar(256) DEFAULT NULL,
  `name` varchar(256) NOT NULL,
  `provider` varchar(256) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_authorities_modules` (`authority_id`),
  CONSTRAINT `fk_authorities_modules` FOREIGN KEY (`authority_id`) REFERENCES `authorities` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO `modules` (`id`, `created_at`, `updated_at`, `authority_id`, `name`, `provider`) VALUES
('0e0fe8e0-e43f-4e8e-bf4a-893e6d3237b3',	'2024-04-19 14:08:33',	'2024-04-19 14:08:33',	'9a50dba6-7ab3-4ee6-8660-8f5901337883',	'kms',	'aws'),
('1f0e1ead-d4b3-4c4b-9313-9fa4a9cf15fe',	'2024-04-19 14:08:34',	'2024-04-19 14:08:34',	'9a50dba6-7ab3-4ee6-8660-8f5901337883',	's3-bucket',	'aws'),
('41cc61e4-e14a-4ed3-a20e-a47160194a53',	'2024-04-19 13:50:05',	'2024-04-19 13:50:05',	'9a50dba6-7ab3-4ee6-8660-8f5901337883',	'rds',	'aws'),
('4bcefa00-6beb-40b9-8740-d13aa6c8c7c3',	'2024-04-19 13:50:02',	'2024-04-19 13:50:02',	'9a50dba6-7ab3-4ee6-8660-8f5901337883',	'cloudfront',	'aws'),
('4fa3edc9-4c2b-42fa-bf43-c9c02240f87d',	'2024-04-19 13:50:00',	'2024-04-19 13:50:01',	'9a50dba6-7ab3-4ee6-8660-8f5901337883',	'security-group',	'aws'),
('51c70b4c-57e1-4795-8f9e-19e7adba3e58',	'2024-04-19 13:50:03',	'2024-04-19 13:50:03',	'9a50dba6-7ab3-4ee6-8660-8f5901337883',	'lambda',	'aws'),
('756e52be-4242-4a2a-9d65-b969f47ff978',	'2024-04-19 13:50:01',	'2024-06-09 21:59:04',	'9a50dba6-7ab3-4ee6-8660-8f5901337883',	'vpc',	'aws'),
('7ea71bb3-63e7-4a20-8f5b-6ec67a241859',	'2024-04-19 13:49:55',	'2024-04-19 13:49:58',	'9a50dba6-7ab3-4ee6-8660-8f5901337883',	'eks',	'aws');

DROP TABLE IF EXISTS `provider_platforms`;
CREATE TABLE `provider_platforms` (
  `id` varchar(256) NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `version_id` varchar(256) DEFAULT NULL,
  `system` varchar(256) NOT NULL,
  `architecture` varchar(256) NOT NULL,
  `location` varchar(256) NOT NULL,
  `sha_sum` varchar(256) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_provider_versions_platforms` (`version_id`),
  CONSTRAINT `fk_provider_versions_platforms` FOREIGN KEY (`version_id`) REFERENCES `provider_versions` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO `provider_platforms` (`id`, `created_at`, `updated_at`, `version_id`, `system`, `architecture`, `location`, `sha_sum`) VALUES
('4a3cfce0-f5ef-4f47-90cd-8238dce05604',	'2024-04-19 14:07:40',	'2024-04-19 14:07:40',	'c51cffe5-9bb3-4d49-b6ff-b567f8fc6e76',	'darwin',	'amd64',	'providers/hashicorp/random/5.46.0/terraform-provider-random_5.46.0_darwin_amd64.zip',	'8c9e8d30c4ef08ee8bcc4294dbf3c2115cd7d9049c6ba21422bd3471d92faf8a'),
('66c87656-e6e8-4058-a6f9-0c0c13eb3e8e',	'2024-04-19 14:08:44',	'2024-04-19 14:08:44',	'32c1b59f-be99-43d6-a370-57b33a6f6204',	'darwin',	'amd64',	'providers/hashicorp/aws/5.46.0/terraform-provider-aws_5.46.0_darwin_amd64.zip',	'90693d936c9a556d2bf945de4920ff82052002eb73139bd7164fafd02920f0ef'),
('73f790c0-a97e-4ecb-a043-78c674abcbcd',	'2024-04-19 14:08:44',	'2024-04-19 14:08:44',	'32c1b59f-be99-43d6-a370-57b33a6f6204',	'linux',	'amd64',	'providers/hashicorp/aws/5.46.0/terraform-provider-aws_5.46.0_linux_amd64.zip',	'37cdf4292649a10f12858622826925e18ad4eca354c31f61d02c66895eb91274'),
('90a002ea-8527-4162-8be5-ad119af55bc4',	'2024-04-19 14:07:40',	'2024-04-19 14:07:40',	'c51cffe5-9bb3-4d49-b6ff-b567f8fc6e76',	'windows',	'amd64',	'providers/hashicorp/random/5.46.0/terraform-provider-random_5.46.0_windows_amd64.zip',	'e4aabf3184bbb556b89e4b195eab1514c86a2914dd01c23ad9813ec17e863a8a'),
('a05d1816-16d6-4b84-8cf7-c7796c49482a',	'2024-04-19 14:08:44',	'2024-04-19 14:08:44',	'32c1b59f-be99-43d6-a370-57b33a6f6204',	'darwin',	'arm64',	'providers/hashicorp/aws/5.46.0/terraform-provider-aws_5.46.0_darwin_arm64.zip',	'e4ede44a112296c9cc77b15e439e41ee15c0e8b3a0dec94ae34df5ebba840e8b'),
('ac164737-f820-4199-8c25-499ce9d8623a',	'2024-04-19 14:07:40',	'2024-04-19 14:07:40',	'c51cffe5-9bb3-4d49-b6ff-b567f8fc6e76',	'darwin',	'arm64',	'providers/hashicorp/random/5.46.0/terraform-provider-random_5.46.0_darwin_arm64.zip',	'b9d1873f14d6033e216510ef541c891f44d249464f13cc07d3f782d09c7d18de'),
('b6e7b3f0-4e52-4b17-80ee-17486aa7bfb2',	'2024-04-19 14:08:44',	'2024-04-19 14:08:44',	'32c1b59f-be99-43d6-a370-57b33a6f6204',	'windows',	'amd64',	'providers/hashicorp/aws/5.46.0/terraform-provider-aws_5.46.0_windows_amd64.zip',	'f2d4de8d8cde69caffede1544ebea74e69fcc4552e1b79ae053519a05c060706'),
('fc0a16f1-df4f-4a39-80be-9469b713ef3d',	'2024-04-19 14:07:40',	'2024-04-19 14:07:40',	'c51cffe5-9bb3-4d49-b6ff-b567f8fc6e76',	'linux',	'amd64',	'providers/hashicorp/random/5.46.0/terraform-provider-random_5.46.0_linux_amd64.zip',	'982542e921970d727ce10ed64795bf36c4dec77a5db0741d4665230d12250a0d');

DROP TABLE IF EXISTS `provider_versions`;
CREATE TABLE `provider_versions` (
  `id` varchar(256) NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `provider_id` varchar(256) DEFAULT NULL,
  `version` varchar(256) NOT NULL,
  `protocols` varchar(256) NOT NULL,
  `sha_sums_url` varchar(256) DEFAULT NULL,
  `sha_sums_signature_url` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_providers_versions` (`provider_id`),
  CONSTRAINT `fk_providers_versions` FOREIGN KEY (`provider_id`) REFERENCES `providers` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO `provider_versions` (`id`, `created_at`, `updated_at`, `provider_id`, `version`, `protocols`, `sha_sums_url`, `sha_sums_signature_url`) VALUES
('32c1b59f-be99-43d6-a370-57b33a6f6204',	'2024-04-19 14:08:44',	'2024-04-19 14:08:44',	'd9ca0a37-363b-48f7-9d9e-2d64d478cc76',	'5.46.0',	'5.0',	'providers/hashicorp/aws/5.46.0/terraform-provider-aws_5.46.0_SHA256SUMS',	'providers/hashicorp/aws/5.46.0/terraform-provider-aws_5.46.0_SHA256SUMS.sig'),
('c51cffe5-9bb3-4d49-b6ff-b567f8fc6e76',	'2024-04-19 14:07:40',	'2024-04-19 14:07:40',	'97376674-323f-4d49-807b-29a291b3340f',	'5.46.0',	'5.0',	'providers/hashicorp/random/5.46.0/terraform-provider-random_5.46.0_SHA256SUMS',	'providers/hashicorp/random/5.46.0/terraform-provider-random_5.46.0_SHA256SUMS.sig');

DROP TABLE IF EXISTS `providers`;
CREATE TABLE `providers` (
  `id` varchar(256) NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `authority_id` varchar(256) DEFAULT NULL,
  `name` varchar(256) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_providers_name` (`name`),
  KEY `fk_authorities_providers` (`authority_id`),
  CONSTRAINT `fk_authorities_providers` FOREIGN KEY (`authority_id`) REFERENCES `authorities` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO `providers` (`id`, `created_at`, `updated_at`, `authority_id`, `name`) VALUES
('97376674-323f-4d49-807b-29a291b3340f',	'2024-04-19 14:07:40',	'2024-04-19 14:07:40',	'04d7980b-9cdd-4cec-bc80-46db639e18b3',	'random'),
('d9ca0a37-363b-48f7-9d9e-2d64d478cc76',	'2024-06-09 21:59:00',	'2024-06-09 21:59:04',	'04d7980b-9cdd-4cec-bc80-46db639e18b3',	'aws');
