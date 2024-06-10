--
-- PostgreSQL database dump
--

-- Dumped from database version 16.2 (Debian 16.2-1.pgdg120+2)
-- Dumped by pg_dump version 16.3

-- Started on 2024-06-09 17:43:52 EEST

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_table_access_method = heap;

--
-- TOC entry 215 (class 1259 OID 16528)
-- Name: authorities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.authorities (
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    name text NOT NULL,
    policy_url text NOT NULL,
    owner text NOT NULL
);


--
-- TOC entry 216 (class 1259 OID 16533)
-- Name: authority_api_keys; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.authority_api_keys (
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    authority_id text,
    expiration timestamp with time zone
);


--
-- TOC entry 217 (class 1259 OID 16538)
-- Name: authority_keys; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.authority_keys (
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    authority_id text,
    key_id text NOT NULL,
    ascii_armor text,
    trust_signature text
);


--
-- TOC entry 218 (class 1259 OID 16543)
-- Name: module_dependencies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.module_dependencies (
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    parent_id text
);


--
-- TOC entry 219 (class 1259 OID 16548)
-- Name: module_providers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.module_providers (
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    parent_id text,
    name text NOT NULL,
    namespace text NOT NULL,
    source text NOT NULL,
    version text NOT NULL
);


--
-- TOC entry 220 (class 1259 OID 16553)
-- Name: module_submodules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.module_submodules (
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    version_id text,
    path text NOT NULL
);


--
-- TOC entry 221 (class 1259 OID 16558)
-- Name: module_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.module_versions (
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    module_id text,
    version text NOT NULL,
    location text NOT NULL
);


--
-- TOC entry 222 (class 1259 OID 16563)
-- Name: modules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.modules (
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    authority_id text,
    name text NOT NULL,
    provider text NOT NULL
);


--
-- TOC entry 223 (class 1259 OID 16568)
-- Name: provider_platforms; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.provider_platforms (
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    version_id text,
    system text NOT NULL,
    architecture text NOT NULL,
    location text NOT NULL,
    sha_sum text NOT NULL
);


--
-- TOC entry 224 (class 1259 OID 16573)
-- Name: provider_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.provider_versions (
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    provider_id text,
    version text NOT NULL,
    protocols text NOT NULL,
    sha_sums_url text,
    sha_sums_signature_url text
);


--
-- TOC entry 225 (class 1259 OID 16578)
-- Name: providers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.providers (
    id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    authority_id text,
    name text NOT NULL
);


--
-- TOC entry 3422 (class 0 OID 16528)
-- Dependencies: 215
-- Data for Name: authorities; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.authorities (id, created_at, updated_at, name, policy_url, owner) VALUES ('9a50dba6-7ab3-4ee6-8660-8f5901337883', '2024-04-19 13:42:02.409412+00', '2024-04-19 13:42:02.409412+00', 'terraform-aws-modules', '', '');
INSERT INTO public.authorities (id, created_at, updated_at, name, policy_url, owner) VALUES ('04d7980b-9cdd-4cec-bc80-46db639e18b3', '2024-04-19 13:42:36.874539+00', '2024-04-19 13:53:57.578055+00', 'hashicorp', 'https://www.hashicorp.com/security.html', '');


--
-- TOC entry 3423 (class 0 OID 16533)
-- Dependencies: 216
-- Data for Name: authority_api_keys; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.authority_api_keys (id, created_at, updated_at, authority_id, expiration) VALUES ('b2eeb8e8-f318-442c-8237-962cad073496', '2024-04-19 13:49:25.257019+00', '2024-04-19 13:49:25.257019+00', '9a50dba6-7ab3-4ee6-8660-8f5901337883', NULL);
INSERT INTO public.authority_api_keys (id, created_at, updated_at, authority_id, expiration) VALUES ('7684e467-4038-4b04-b03b-eec76521c7c2', '2024-04-19 13:51:18.950561+00', '2024-04-19 13:51:18.950561+00', '04d7980b-9cdd-4cec-bc80-46db639e18b3', NULL);


--
-- TOC entry 3424 (class 0 OID 16538)
-- Dependencies: 217
-- Data for Name: authority_keys; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.authority_keys (id, created_at, updated_at, authority_id, key_id, ascii_armor, trust_signature) VALUES ('df11cb4a-b8c7-4b32-9802-b70fcdda729f', '2024-04-19 13:53:57.579779+00', '2024-04-19 13:53:57.579779+00', '04d7980b-9cdd-4cec-bc80-46db639e18b3', '34365D9472D7468F', '-----BEGIN PGP PUBLIC KEY BLOCK-----

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
ZF5q4h4I33PSGDdSvGXn9UMY5Isjpg==
=7pIB
-----END PGP PUBLIC KEY BLOCK-----', '');


--
-- TOC entry 3425 (class 0 OID 16543)
-- Dependencies: 218
-- Data for Name: module_dependencies; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- TOC entry 3426 (class 0 OID 16548)
-- Dependencies: 219
-- Data for Name: module_providers; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- TOC entry 3427 (class 0 OID 16553)
-- Dependencies: 220
-- Data for Name: module_submodules; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- TOC entry 3428 (class 0 OID 16558)
-- Dependencies: 221
-- Data for Name: module_versions; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.module_versions (id, created_at, updated_at, module_id, version, location) VALUES ('8a37151b-9adb-48f3-bd7e-baad8d8308ad', '2024-04-19 14:08:33.710431+00', '2024-04-19 14:08:33.710431+00', '0e0fe8e0-e43f-4e8e-bf4a-893e6d3237b3', '2.2.1', 'modules/terraform-aws-modules/kms/aws/2.2.1.zip');
INSERT INTO public.module_versions (id, created_at, updated_at, module_id, version, location) VALUES ('017b725a-3200-4c36-8020-0742223429a7', '2024-04-19 14:08:34.414356+00', '2024-04-19 14:08:34.414356+00', '1f0e1ead-d4b3-4c4b-9313-9fa4a9cf15fe', '4.1.2', 'modules/terraform-aws-modules/s3-bucket/aws/4.1.2.zip');
INSERT INTO public.module_versions (id, created_at, updated_at, module_id, version, location) VALUES ('7145990b-3422-4ac0-9b64-782cd50a8494', '2024-04-19 13:49:52.623902+00', '2024-04-19 13:49:52.623902+00', '756e52be-4242-4a2a-9d65-b969f47ff978', '5.5.3', 'modules/terraform-aws-modules/vpc/aws/5.5.3.zip');
INSERT INTO public.module_versions (id, created_at, updated_at, module_id, version, location) VALUES ('0c8e4b8c-f025-4560-9803-7b8c519c279e', '2024-04-19 13:49:53.43125+00', '2024-04-19 13:49:53.43125+00', '756e52be-4242-4a2a-9d65-b969f47ff978', '5.6.0', 'modules/terraform-aws-modules/vpc/aws/5.6.0.zip');
INSERT INTO public.module_versions (id, created_at, updated_at, module_id, version, location) VALUES ('5bd0c46b-09db-481f-9211-e7db7a883122', '2024-04-19 13:49:54.255857+00', '2024-04-19 13:49:54.255857+00', '756e52be-4242-4a2a-9d65-b969f47ff978', '5.7.0', 'modules/terraform-aws-modules/vpc/aws/5.7.0.zip');
INSERT INTO public.module_versions (id, created_at, updated_at, module_id, version, location) VALUES ('98f8afdc-7c04-4460-b244-fa520e76a6f2', '2024-04-19 13:49:54.895768+00', '2024-04-19 13:49:54.895768+00', '756e52be-4242-4a2a-9d65-b969f47ff978', '5.7.1', 'modules/terraform-aws-modules/vpc/aws/5.7.1.zip');
INSERT INTO public.module_versions (id, created_at, updated_at, module_id, version, location) VALUES ('602d385d-7ca3-46e1-bbfc-c61b2dca69f0', '2024-04-19 13:49:55.859084+00', '2024-04-19 13:49:55.859084+00', '7ea71bb3-63e7-4a20-8f5b-6ec67a241859', '20.8.0', 'modules/terraform-aws-modules/eks/aws/20.8.0.zip');
INSERT INTO public.module_versions (id, created_at, updated_at, module_id, version, location) VALUES ('4b272992-e913-4aa2-9164-d8962a3a6489', '2024-04-19 13:49:56.767176+00', '2024-04-19 13:49:56.767176+00', '7ea71bb3-63e7-4a20-8f5b-6ec67a241859', '20.8.1', 'modules/terraform-aws-modules/eks/aws/20.8.1.zip');
INSERT INTO public.module_versions (id, created_at, updated_at, module_id, version, location) VALUES ('6ad6ac12-89e6-4ec7-b322-493f85f2f7ad', '2024-04-19 13:49:57.857993+00', '2024-04-19 13:49:57.857993+00', '7ea71bb3-63e7-4a20-8f5b-6ec67a241859', '20.8.2', 'modules/terraform-aws-modules/eks/aws/20.8.2.zip');
INSERT INTO public.module_versions (id, created_at, updated_at, module_id, version, location) VALUES ('3049393a-3c01-4870-8898-5fc61f6d3d6f', '2024-04-19 13:49:58.756136+00', '2024-04-19 13:49:58.756136+00', '7ea71bb3-63e7-4a20-8f5b-6ec67a241859', '20.8.3', 'modules/terraform-aws-modules/eks/aws/20.8.3.zip');
INSERT INTO public.module_versions (id, created_at, updated_at, module_id, version, location) VALUES ('a11de7e8-eb30-40f4-9fb4-d88582323cc9', '2024-04-19 13:50:00.286046+00', '2024-04-19 13:50:00.286046+00', '4fa3edc9-4c2b-42fa-bf43-c9c02240f87d', '5.1.1', 'modules/terraform-aws-modules/security-group/aws/5.1.1.zip');
INSERT INTO public.module_versions (id, created_at, updated_at, module_id, version, location) VALUES ('85b5ccad-79fd-4413-a61e-4df562fd4084', '2024-04-19 13:50:01.674034+00', '2024-04-19 13:50:01.674034+00', '4fa3edc9-4c2b-42fa-bf43-c9c02240f87d', '5.1.2', 'modules/terraform-aws-modules/security-group/aws/5.1.2.zip');
INSERT INTO public.module_versions (id, created_at, updated_at, module_id, version, location) VALUES ('f070736f-2761-4b4a-a789-0593170a382b', '2024-04-19 13:50:02.365267+00', '2024-04-19 13:50:02.365267+00', '4bcefa00-6beb-40b9-8740-d13aa6c8c7c3', '3.4.0', 'modules/terraform-aws-modules/cloudfront/aws/3.4.0.zip');
INSERT INTO public.module_versions (id, created_at, updated_at, module_id, version, location) VALUES ('1b8cee2b-5c72-4904-86a9-63a42368c40c', '2024-04-19 13:50:03.352278+00', '2024-04-19 13:50:03.352278+00', '51c70b4c-57e1-4795-8f9e-19e7adba3e58', '7.2.6', 'modules/terraform-aws-modules/lambda/aws/7.2.6.zip');
INSERT INTO public.module_versions (id, created_at, updated_at, module_id, version, location) VALUES ('17647521-e531-4a83-b690-cb43ac87ba20', '2024-04-19 13:50:05.386202+00', '2024-04-19 13:50:05.386202+00', '41cc61e4-e14a-4ed3-a20e-a47160194a53', '6.5.4', 'modules/terraform-aws-modules/rds/aws/6.5.4.zip');


--
-- TOC entry 3429 (class 0 OID 16563)
-- Dependencies: 222
-- Data for Name: modules; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.modules (id, created_at, updated_at, authority_id, name, provider) VALUES ('756e52be-4242-4a2a-9d65-b969f47ff978', '2024-04-19 13:49:52.622925+00', '2024-04-19 13:49:54.894851+00', '9a50dba6-7ab3-4ee6-8660-8f5901337883', 'vpc', 'aws');
INSERT INTO public.modules (id, created_at, updated_at, authority_id, name, provider) VALUES ('7ea71bb3-63e7-4a20-8f5b-6ec67a241859', '2024-04-19 13:49:55.8582+00', '2024-04-19 13:49:58.755247+00', '9a50dba6-7ab3-4ee6-8660-8f5901337883', 'eks', 'aws');
INSERT INTO public.modules (id, created_at, updated_at, authority_id, name, provider) VALUES ('4fa3edc9-4c2b-42fa-bf43-c9c02240f87d', '2024-04-19 13:50:00.285191+00', '2024-04-19 13:50:01.67313+00', '9a50dba6-7ab3-4ee6-8660-8f5901337883', 'security-group', 'aws');
INSERT INTO public.modules (id, created_at, updated_at, authority_id, name, provider) VALUES ('4bcefa00-6beb-40b9-8740-d13aa6c8c7c3', '2024-04-19 13:50:02.364399+00', '2024-04-19 13:50:02.364399+00', '9a50dba6-7ab3-4ee6-8660-8f5901337883', 'cloudfront', 'aws');
INSERT INTO public.modules (id, created_at, updated_at, authority_id, name, provider) VALUES ('51c70b4c-57e1-4795-8f9e-19e7adba3e58', '2024-04-19 13:50:03.351386+00', '2024-04-19 13:50:03.351386+00', '9a50dba6-7ab3-4ee6-8660-8f5901337883', 'lambda', 'aws');
INSERT INTO public.modules (id, created_at, updated_at, authority_id, name, provider) VALUES ('41cc61e4-e14a-4ed3-a20e-a47160194a53', '2024-04-19 13:50:05.385336+00', '2024-04-19 13:50:05.385336+00', '9a50dba6-7ab3-4ee6-8660-8f5901337883', 'rds', 'aws');
INSERT INTO public.modules (id, created_at, updated_at, authority_id, name, provider) VALUES ('0e0fe8e0-e43f-4e8e-bf4a-893e6d3237b3', '2024-04-19 14:08:33.709555+00', '2024-04-19 14:08:33.709555+00', '9a50dba6-7ab3-4ee6-8660-8f5901337883', 'kms', 'aws');
INSERT INTO public.modules (id, created_at, updated_at, authority_id, name, provider) VALUES ('1f0e1ead-d4b3-4c4b-9313-9fa4a9cf15fe', '2024-04-19 14:08:34.413827+00', '2024-04-19 14:08:34.413827+00', '9a50dba6-7ab3-4ee6-8660-8f5901337883', 's3-bucket', 'aws');


--
-- TOC entry 3430 (class 0 OID 16568)
-- Dependencies: 223
-- Data for Name: provider_platforms; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.provider_platforms (id, created_at, updated_at, version_id, system, architecture, location, sha_sum) VALUES ('fc0a16f1-df4f-4a39-80be-9469b713ef3d', '2024-04-19 14:07:40.0566+00', '2024-04-19 14:07:40.0566+00', 'c51cffe5-9bb3-4d49-b6ff-b567f8fc6e76', 'linux', 'amd64', 'providers/hashicorp/random/5.46.0/terraform-provider-random_5.46.0_linux_amd64.zip', '982542e921970d727ce10ed64795bf36c4dec77a5db0741d4665230d12250a0d');
INSERT INTO public.provider_platforms (id, created_at, updated_at, version_id, system, architecture, location, sha_sum) VALUES ('4a3cfce0-f5ef-4f47-90cd-8238dce05604', '2024-04-19 14:07:40.0566+00', '2024-04-19 14:07:40.0566+00', 'c51cffe5-9bb3-4d49-b6ff-b567f8fc6e76', 'darwin', 'amd64', 'providers/hashicorp/random/5.46.0/terraform-provider-random_5.46.0_darwin_amd64.zip', '8c9e8d30c4ef08ee8bcc4294dbf3c2115cd7d9049c6ba21422bd3471d92faf8a');
INSERT INTO public.provider_platforms (id, created_at, updated_at, version_id, system, architecture, location, sha_sum) VALUES ('ac164737-f820-4199-8c25-499ce9d8623a', '2024-04-19 14:07:40.0566+00', '2024-04-19 14:07:40.0566+00', 'c51cffe5-9bb3-4d49-b6ff-b567f8fc6e76', 'darwin', 'arm64', 'providers/hashicorp/random/5.46.0/terraform-provider-random_5.46.0_darwin_arm64.zip', 'b9d1873f14d6033e216510ef541c891f44d249464f13cc07d3f782d09c7d18de');
INSERT INTO public.provider_platforms (id, created_at, updated_at, version_id, system, architecture, location, sha_sum) VALUES ('90a002ea-8527-4162-8be5-ad119af55bc4', '2024-04-19 14:07:40.0566+00', '2024-04-19 14:07:40.0566+00', 'c51cffe5-9bb3-4d49-b6ff-b567f8fc6e76', 'windows', 'amd64', 'providers/hashicorp/random/5.46.0/terraform-provider-random_5.46.0_windows_amd64.zip', 'e4aabf3184bbb556b89e4b195eab1514c86a2914dd01c23ad9813ec17e863a8a');
INSERT INTO public.provider_platforms (id, created_at, updated_at, version_id, system, architecture, location, sha_sum) VALUES ('73f790c0-a97e-4ecb-a043-78c674abcbcd', '2024-04-19 14:08:44.545558+00', '2024-04-19 14:08:44.545558+00', '32c1b59f-be99-43d6-a370-57b33a6f6204', 'linux', 'amd64', 'providers/hashicorp/aws/5.46.0/terraform-provider-aws_5.46.0_linux_amd64.zip', '37cdf4292649a10f12858622826925e18ad4eca354c31f61d02c66895eb91274');
INSERT INTO public.provider_platforms (id, created_at, updated_at, version_id, system, architecture, location, sha_sum) VALUES ('66c87656-e6e8-4058-a6f9-0c0c13eb3e8e', '2024-04-19 14:08:44.545558+00', '2024-04-19 14:08:44.545558+00', '32c1b59f-be99-43d6-a370-57b33a6f6204', 'darwin', 'amd64', 'providers/hashicorp/aws/5.46.0/terraform-provider-aws_5.46.0_darwin_amd64.zip', '90693d936c9a556d2bf945de4920ff82052002eb73139bd7164fafd02920f0ef');
INSERT INTO public.provider_platforms (id, created_at, updated_at, version_id, system, architecture, location, sha_sum) VALUES ('a05d1816-16d6-4b84-8cf7-c7796c49482a', '2024-04-19 14:08:44.545558+00', '2024-04-19 14:08:44.545558+00', '32c1b59f-be99-43d6-a370-57b33a6f6204', 'darwin', 'arm64', 'providers/hashicorp/aws/5.46.0/terraform-provider-aws_5.46.0_darwin_arm64.zip', 'e4ede44a112296c9cc77b15e439e41ee15c0e8b3a0dec94ae34df5ebba840e8b');
INSERT INTO public.provider_platforms (id, created_at, updated_at, version_id, system, architecture, location, sha_sum) VALUES ('b6e7b3f0-4e52-4b17-80ee-17486aa7bfb2', '2024-04-19 14:08:44.545558+00', '2024-04-19 14:08:44.545558+00', '32c1b59f-be99-43d6-a370-57b33a6f6204', 'windows', 'amd64', 'providers/hashicorp/aws/5.46.0/terraform-provider-aws_5.46.0_windows_amd64.zip', 'f2d4de8d8cde69caffede1544ebea74e69fcc4552e1b79ae053519a05c060706');


--
-- TOC entry 3431 (class 0 OID 16573)
-- Dependencies: 224
-- Data for Name: provider_versions; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.provider_versions (id, created_at, updated_at, provider_id, version, protocols, sha_sums_url, sha_sums_signature_url) VALUES ('c51cffe5-9bb3-4d49-b6ff-b567f8fc6e76', '2024-04-19 14:07:40.05594+00', '2024-04-19 14:07:40.05594+00', '97376674-323f-4d49-807b-29a291b3340f', '5.46.0', '5.0', 'providers/hashicorp/random/5.46.0/terraform-provider-random_5.46.0_SHA256SUMS', 'providers/hashicorp/random/5.46.0/terraform-provider-random_5.46.0_SHA256SUMS.sig');
INSERT INTO public.provider_versions (id, created_at, updated_at, provider_id, version, protocols, sha_sums_url, sha_sums_signature_url) VALUES ('32c1b59f-be99-43d6-a370-57b33a6f6204', '2024-04-19 14:08:44.545057+00', '2024-04-19 14:08:44.545057+00', 'd9ca0a37-363b-48f7-9d9e-2d64d478cc76', '5.46.0', '5.0', 'providers/hashicorp/aws/5.46.0/terraform-provider-aws_5.46.0_SHA256SUMS', 'providers/hashicorp/aws/5.46.0/terraform-provider-aws_5.46.0_SHA256SUMS.sig');


--
-- TOC entry 3432 (class 0 OID 16578)
-- Dependencies: 225
-- Data for Name: providers; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.providers (id, created_at, updated_at, authority_id, name) VALUES ('97376674-323f-4d49-807b-29a291b3340f', '2024-04-19 14:07:40.054966+00', '2024-04-19 14:07:40.054966+00', '04d7980b-9cdd-4cec-bc80-46db639e18b3', 'random');
INSERT INTO public.providers (id, created_at, updated_at, authority_id, name) VALUES ('d9ca0a37-363b-48f7-9d9e-2d64d478cc76', '2024-04-19 14:08:44.54442+00', '2024-04-19 14:08:44.54442+00', '04d7980b-9cdd-4cec-bc80-46db639e18b3', 'aws');


--
-- TOC entry 3243 (class 2606 OID 16585)
-- Name: authorities authorities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.authorities
    ADD CONSTRAINT authorities_pkey PRIMARY KEY (id);


--
-- TOC entry 3247 (class 2606 OID 16587)
-- Name: authority_api_keys authority_api_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.authority_api_keys
    ADD CONSTRAINT authority_api_keys_pkey PRIMARY KEY (id);


--
-- TOC entry 3249 (class 2606 OID 16589)
-- Name: authority_keys authority_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.authority_keys
    ADD CONSTRAINT authority_keys_pkey PRIMARY KEY (id);


--
-- TOC entry 3251 (class 2606 OID 16591)
-- Name: module_dependencies module_dependencies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_dependencies
    ADD CONSTRAINT module_dependencies_pkey PRIMARY KEY (id);


--
-- TOC entry 3253 (class 2606 OID 16593)
-- Name: module_providers module_providers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_providers
    ADD CONSTRAINT module_providers_pkey PRIMARY KEY (id);


--
-- TOC entry 3255 (class 2606 OID 16595)
-- Name: module_submodules module_submodules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_submodules
    ADD CONSTRAINT module_submodules_pkey PRIMARY KEY (id);


--
-- TOC entry 3257 (class 2606 OID 16597)
-- Name: module_versions module_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_versions
    ADD CONSTRAINT module_versions_pkey PRIMARY KEY (id);


--
-- TOC entry 3259 (class 2606 OID 16599)
-- Name: modules modules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.modules
    ADD CONSTRAINT modules_pkey PRIMARY KEY (id);


--
-- TOC entry 3261 (class 2606 OID 16601)
-- Name: provider_platforms provider_platforms_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.provider_platforms
    ADD CONSTRAINT provider_platforms_pkey PRIMARY KEY (id);


--
-- TOC entry 3263 (class 2606 OID 16603)
-- Name: provider_versions provider_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.provider_versions
    ADD CONSTRAINT provider_versions_pkey PRIMARY KEY (id);


--
-- TOC entry 3266 (class 2606 OID 16605)
-- Name: providers providers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providers
    ADD CONSTRAINT providers_pkey PRIMARY KEY (id);


--
-- TOC entry 3244 (class 1259 OID 16606)
-- Name: idx_authorities_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_authorities_name ON public.authorities USING btree (name);


--
-- TOC entry 3245 (class 1259 OID 16607)
-- Name: idx_authorities_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_authorities_owner ON public.authorities USING btree (owner);


--
-- TOC entry 3264 (class 1259 OID 16608)
-- Name: idx_providers_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_providers_name ON public.providers USING btree (name);


--
-- TOC entry 3267 (class 2606 OID 16609)
-- Name: authority_api_keys fk_authorities_api_keys; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.authority_api_keys
    ADD CONSTRAINT fk_authorities_api_keys FOREIGN KEY (authority_id) REFERENCES public.authorities(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3268 (class 2606 OID 16614)
-- Name: authority_keys fk_authorities_keys; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.authority_keys
    ADD CONSTRAINT fk_authorities_keys FOREIGN KEY (authority_id) REFERENCES public.authorities(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3275 (class 2606 OID 16619)
-- Name: modules fk_authorities_modules; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.modules
    ADD CONSTRAINT fk_authorities_modules FOREIGN KEY (authority_id) REFERENCES public.authorities(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3278 (class 2606 OID 16624)
-- Name: providers fk_authorities_providers; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providers
    ADD CONSTRAINT fk_authorities_providers FOREIGN KEY (authority_id) REFERENCES public.authorities(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3269 (class 2606 OID 16629)
-- Name: module_dependencies fk_module_submodules_dependencies; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_dependencies
    ADD CONSTRAINT fk_module_submodules_dependencies FOREIGN KEY (parent_id) REFERENCES public.module_submodules(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3271 (class 2606 OID 16634)
-- Name: module_providers fk_module_submodules_providers; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_providers
    ADD CONSTRAINT fk_module_submodules_providers FOREIGN KEY (parent_id) REFERENCES public.module_submodules(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3270 (class 2606 OID 16639)
-- Name: module_dependencies fk_module_versions_dependencies; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_dependencies
    ADD CONSTRAINT fk_module_versions_dependencies FOREIGN KEY (parent_id) REFERENCES public.module_versions(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3272 (class 2606 OID 16644)
-- Name: module_providers fk_module_versions_providers; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_providers
    ADD CONSTRAINT fk_module_versions_providers FOREIGN KEY (parent_id) REFERENCES public.module_versions(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3273 (class 2606 OID 16649)
-- Name: module_submodules fk_module_versions_submodules; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_submodules
    ADD CONSTRAINT fk_module_versions_submodules FOREIGN KEY (version_id) REFERENCES public.module_versions(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3274 (class 2606 OID 16654)
-- Name: module_versions fk_modules_versions; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_versions
    ADD CONSTRAINT fk_modules_versions FOREIGN KEY (module_id) REFERENCES public.modules(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3276 (class 2606 OID 16659)
-- Name: provider_platforms fk_provider_versions_platforms; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.provider_platforms
    ADD CONSTRAINT fk_provider_versions_platforms FOREIGN KEY (version_id) REFERENCES public.provider_versions(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3277 (class 2606 OID 16664)
-- Name: provider_versions fk_providers_versions; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.provider_versions
    ADD CONSTRAINT fk_providers_versions FOREIGN KEY (provider_id) REFERENCES public.providers(id) ON UPDATE CASCADE ON DELETE CASCADE;


-- Completed on 2024-06-09 17:43:52 EEST

--
-- PostgreSQL database dump complete
--

