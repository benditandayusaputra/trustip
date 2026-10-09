# Daftar periksa keamanan

Setiap aturan `R-SEC` di `03-ATURAN-REKAYASA.md` bagian 7, dengan status dan bukti. Bukti berupa uji otomatis atau pemeriksaan mesin yang berjalan di `make check`, `make test`, atau `make e2e`.

Status:

- **Terpenuhi**: ada uji atau pemeriksaan mesin yang membuktikannya.
- **Terpenuhi, tanpa uji khusus**: perilakunya ada di kode yang disebut, tetapi belum ada uji yang khusus untuknya.
- **Operasional**: dibuktikan oleh berkas deploy atau prosedur, bukan uji.
- **Belum**: belum ada. Alasannya ditulis di bagian akhir.

## 7.2 Kata sandi

| Aturan | Ringkas | Status | Bukti |
|---|---|---|---|
| R-SEC-01 | Sandi NFKC, 10 sampai 128, tolak umum dan identitas | Terpenuhi | `api/internal/modules/auth/domain/password_test.go` `TestPasswordPolicy_RSEC01_LengthBoundaries`, `TestPasswordPolicy_RSEC01_PasswordsWithIdentityAreRejected`; `api/internal/platform/httpserver/decode_test.go` `TestNormalizePassword_RSEC01_AppliesNFKC` |
| R-SEC-02 | Pepper HMAC sebelum Argon2id | Terpenuhi | `api/internal/platform/security/password_test.go` `TestVerify_BRAUTH04_WrongPepperDoesNotMatch` |
| R-SEC-03 | Argon2id, salt 16 byte, parameter minimum | Terpenuhi | `api/internal/platform/security/password_test.go` `TestNewPasswordHasher_RSEC03_WeakerThanMinimumIsRejected`, `TestHash_BRAUTH04_SaltDiffersForTheSamePassword` |
| R-SEC-04 | Simpan string PHC dan versi pepper | Terpenuhi | `api/internal/platform/security/password_test.go` `TestHash_RSEC04_StoresPHCStringWithParamsAndPepperVersion` |
| R-SEC-05 | Pepper hanya di lingkungan, bisa dirotasi | Terpenuhi | `api/internal/platform/security/password_test.go` `TestVerify_RSEC05_RotatedPepperVerifiesOldHashAndAsksForRehash`; `api/internal/modules/auth/service/login_test.go` `TestLogin_RSEC05_OutdatedHashIsReplacedAfterSuccess` |
| R-SEC-06 | Waktu tetap, hash tiruan untuk email asing | Terpenuhi | `api/internal/modules/auth/service/login_test.go` `TestLogin_RSEC06_UnknownEmailStillSpendsOneHash` |
| R-SEC-07 | Maksimal 4 hash bersamaan, lalu 503 | Terpenuhi | `api/internal/platform/security/password_internal_test.go` `TestHash_RSEC07_FifthConcurrentHashIsUnavailable`; `api/internal/modules/auth/service/register_test.go` `TestRegister_RSEC07_BusyHasherIsUnavailable` |
| R-SEC-08 | Pesan gagal masuk tidak membedakan | Terpenuhi | `api/cmd/server/login_test.go` `TestLogin_RSEC08_WrongPasswordAndUnknownEmailLookTheSame`; `app/src/lib/features/auth/login.svelte.test.ts` |

## 7.3 Captcha

| Aturan | Ringkas | Status | Bukti |
|---|---|---|---|
| R-SEC-09 | Captcha di daftar, masuk, akun demo | Terpenuhi | `api/internal/modules/auth/service/register_test.go` `TestRegister_BRAUTH05_FailedCaptchaStopsBeforeAnythingElse`; `api/internal/modules/auth/service/login_test.go` `TestLogin_BRAUTH05_InvalidCaptchaIsRejectedWithoutCheckingThePassword`. Tombol akun demo mengisi formulir masuk yang sama (ADR 0044). Lupa kata sandi adalah P1 |
| R-SEC-10 | Verifikasi di server: sukses, host, aksi | Terpenuhi | `api/internal/platform/captcha/turnstile_test.go` `TestVerify_RSEC10_MatchingHostnameAndActionPasses` |
| R-SEC-11 | Gagal atau tak dijawab berarti `CAPTCHA_FAILED` | Terpenuhi | `api/internal/platform/captcha/turnstile_test.go` `TestVerify_RSEC11_UnreachableServiceFails`, `TestVerify_RSEC11_SilentServiceFailsAfterTimeout` |
| R-SEC-12 | Port tiruan, tanpa sakelar di produksi | Terpenuhi | `api/internal/platform/captcha/turnstile_test.go` `TestVerify_RSEC12_OfficialTestingKeyFailureStillFails`; `api/internal/platform/config/config_test.go` `TestLoad_ProductionWithTurnstileTestKey_IsRejected` |
| R-SEC-13 | Asal widget di CSP, disebut di privasi | Terpenuhi | `app/csp.config.test.ts`; `app/src/lib/features/legal/legal.svelte.test.ts` |

## 7.4 Sesi

| Aturan | Ringkas | Status | Bukti |
|---|---|---|---|
| R-SEC-14 | Token hanya di cookie server | Terpenuhi | `api/cmd/server/login_test.go` `TestLogin_RSEC14_TokensLiveOnlyInHttpOnlyCookies` |
| R-SEC-15 | Awalan `__Host-` di produksi | Terpenuhi | `api/internal/modules/auth/handler/cookies_test.go` `TestCookiePolicy_RSEC15_ProductionCookiesAreBoundToOneHost`, `TestCookiePolicy_RSEC15_LocalCookiesDropThePrefixAndSecure` |
| R-SEC-16 | JWT HS256 dengan `kid` | Terpenuhi | `api/internal/platform/security/token_test.go` `TestVerify_RSEC16_Rejections`, `TestSign_RSEC16_MatchesAnIndependentHS256Implementation` |
| R-SEC-17 | Token penyegar di-hash, diputar, pakai ulang mencabut | Terpenuhi | `api/cmd/server/session_test.go` `TestRefresh_RSEC17_ReusingARotatedTokenRevokesTheWholeFamily`; `api/internal/modules/auth/repository/credentials_test.go` `TestCreateSession_RSEC17_StoresOnlyTheTokenHash` |
| R-SEC-18 | Tenggang 10 detik untuk dua tab | Terpenuhi | `api/cmd/server/session_integration_test.go` `TestRefresh_RSEC18_ConcurrentRefreshesOfOneTokenAgreeOnOnePair`; `api/cmd/server/session_test.go` `TestRefresh_RSEC18_TwoTabsWithinTenSecondsGetTheSamePair` |
| R-SEC-19 | Keluar dan keluar dari semua perangkat | Terpenuhi | `api/cmd/server/session_integration_test.go` `TestLogOutEverywhere_RSEC19_RevokesRealSessionsAcrossFamilies`; `api/cmd/server/account_test.go` `TestMe_RSEC19_AccessTokenStopsWorkingAfterLogOutEverywhere` |
| R-SEC-20 | Ganti peran, nonaktif, tangguh mencabut sesi | Belum | Lihat catatan di bawah |

## 7.5 dan 7.6 Lintas situs dan tanda tangan tepi

| Aturan | Ringkas | Status | Bukti |
|---|---|---|---|
| R-SEC-21 | SameSite, `Origin`, `X-Trustip-Client` | Terpenuhi | `api/cmd/server/account_test.go` `TestRequireSameOrigin_RSEC21_WritesFromAnotherSiteAreRejected`; `app/tests/xss.test.ts` "R-SEC-21 a write sent from another site is refused" |
| R-SEC-22 | Isi tulis hanya JSON atau multipart | Terpenuhi | `api/cmd/server/account_test.go` `TestRequireSameOrigin_RSEC22_WriteBodyMustBeJSON`; `api/internal/platform/httpserver/decode_test.go` `TestDecodeJSON_RSEC22_Layer1Rejections` |
| R-SEC-23 | Tanda tangan tepi: kunci, waktu, nonce | Terpenuhi | `api/internal/platform/security/edge_test.go` `TestVerify_Rejections`, `TestVerify_ReusedNonceWithin120Seconds_IsRejected`; `api/internal/platform/httpserver/server_test.go` `TestEdgeSignature_UnsignedRequest_Returns401AndLogsSecurityEvent` |
| R-SEC-24 | IP klien ikut ditandatangani | Terpenuhi | `api/internal/platform/security/edge_test.go` `TestVerify_Rejections`; `api/internal/platform/httpserver/server_test.go` `TestAccessLog_UsesSignedClientIP` |
| R-SEC-25 | Kunci HMAC 32 byte berversi | Terpenuhi | `api/internal/platform/config/config_test.go` `TestLoad_ShortSecret_IsRejected`; `app/src/lib/server/edge-signer.test.ts` |
| R-SEC-26 | Tiga jalur di luar proxy | Terpenuhi | `api/internal/platform/sse/tickets_test.go` `TestIssue_RSEC26_TicketLivesThirtySeconds`, `TestRedeem_RSEC26_TicketWorksOnce`; `api/internal/platform/httpserver/server_test.go` `TestEdgeSignature_Healthz_NeedsNoSignature` |

## 7.7 Segel

| Aturan | Ringkas | Status | Bukti |
|---|---|---|---|
| R-SEC-27 | Peristiwa dagang yang disegel | Terpenuhi | `api/internal/modules/settlement/domain/document_test.go` `TestSealRequest_RSEC27_IssueIsSealedAsASystemEvent`; `api/internal/modules/partnership/service/ending_test.go` `TestEndPartnership_BRPRT03_SettlesFirstThenEndsSealsAndTellsThePartner` |
| R-SEC-28 | Bentuk kanonik dan SHA-256 | Terpenuhi | `api/internal/kernel/canonical/canonical_test.go` `TestEncode_RSEC28_SortedKeysWithoutWhitespace`, `TestDigest_RSEC28_SHA256OfTheCanonicalForm` |
| R-SEC-29 | Entri berantai bertanda tangan Ed25519 | Terpenuhi | `api/internal/platform/security/seal_test.go` `TestSealer_RSEC29_SharedSealVectorSignsTheSame`, `TestSealer_RSEC29_ChangedDigestFailsVerification` |
| R-SEC-30 | Segel dalam transaksi, rantai tak bercabang | Terpenuhi | `api/internal/modules/seal/repository/seals_test.go` `TestSeal_RSEC30_ConcurrentEventsDoNotForkTheChain` |
| R-SEC-31 | `seals` hanya bisa ditambah | Terpenuhi | `api/internal/modules/seal/repository/seals_test.go` `TestSeals_RSEC31_UpdateAndDeleteAreRejected`; `api/cmd/server/demo_integration_test.go` `TestDemoSeed_Spec25_ResetRebuildsThirtyDaysOfDemoHistory` (hapus hanya untuk demo) |
| R-SEC-32 | Kunci privat di lingkungan, kunci publik terbit | Terpenuhi | `api/internal/platform/security/seal_test.go` `TestPublicKeys_RSEC32_PublishesEveryKeyOfTheRing`; `api/cmd/server/seal_verification_integration_test.go` `TestSealKeys_FRSIG04_PublicKeyIsPublished` |
| R-SEC-33 | Hitung ulang hash, periksa rantai harian | Terpenuhi | `api/cmd/server/agreements_integration_test.go` `TestAgreement_RSEC33_PriceChangedInTheDatabaseShowsAnInvalidSeal`; `api/internal/modules/seal/service/public_test.go` `TestCheckChains_FRSIG05_ReportsTheBrokenPartnership` |
| R-SEC-34 | Verifikasi publik tanpa nilai uang | Terpenuhi | `api/cmd/server/seal_verification_integration_test.go` `TestVerifySeal_FRSIG03_RSEC34_VisitorSeesTheSealWithoutAmounts` |
| R-SEC-35 | Hash bukti bayar ikut disegel | Terpenuhi | `api/internal/modules/settlement/domain/payment_test.go` `TestRecord_RSEC35_SealedContentLocksTheProofHash` |

## 7.8 Hak akses

| Aturan | Ringkas | Status | Bukti |
|---|---|---|---|
| R-SEC-36 | `policy.Require` di baris pertama | Terpenuhi | `api/internal/policy_check_test.go` `TestArchitecture_RSEC36_EveryUseCaseRequiresAPolicyFirst` |
| R-SEC-37 | Tabel izin sama dengan spesifikasi | Terpenuhi | `api/internal/kernel/policy/spec_test.go` `TestTable_RSEC37_MatchesTheRoleTableInTheSpecRowByRow` |
| R-SEC-38 | Lingkup data di SQL | Terpenuhi | `api/internal/modules/product/repository/products_test.go` `TestGetProduct_RSEC38_ProductOfAnotherConsignorIsNotFound`; `api/internal/modules/file/repository/files_test.go` `TestGetFile_RSEC38_FileOfAnotherOrganizationIsNotFound` |
| R-SEC-39 | Organisasi lain 404, peran kurang 403 | Terpenuhi | `api/cmd/server/isolation_integration_test.go` `TestIsolation_RSEC41_OtherOrganizationsGetNotFoundOnEveryRouteWithAnID`; `api/internal/kernel/policy/policy_test.go` `TestRequire_BRAUTHZ01_RoleWithoutGrantIsForbidden` |
| R-SEC-40 | Tampilan eksplisit per pemirsa | Terpenuhi | `api/cmd/server/public_profiles_integration_test.go` `TestPublicConsignor_FRPUB03_BRAUTHZ04_VisitorSeesProductsWithoutPrivateDetails` |
| R-SEC-41 | Uji isolasi untuk setiap rute ber-ID | Terpenuhi | `api/cmd/server/isolation_test.go` `TestIsolation_RSEC41_EveryRouteWithAnIDHasAnIsolationCase` |
| R-SEC-42 | Tidak mengklaim keamanan tingkat baris | Terpenuhi, tanpa uji khusus | Tidak ada RLS di skema maupun klaimnya di dokumen |

## 7.9 Pembatasan laju

| Aturan | Ringkas | Status | Bukti |
|---|---|---|---|
| R-SEC-43 | 429 baku dengan `Retry-After` dan kepala `RateLimit` | Terpenuhi | `api/internal/platform/httpserver/server_test.go` `TestRateLimit_EveryResponseCarriesRateLimitHeaders`, `TestRateLimit_OverLimit_Returns429WithRetryAfterAndLogsEvent` |
| R-SEC-44 | Batas bisnis di basis data | Terpenuhi | `api/internal/modules/proposal/repository/direct_test.go` `TestConsumeQuota_BRPRP06_RSEC44_DatabaseStopsTheEleventhProposalOfTheDay`; `api/internal/modules/need/repository/needs_test.go` `TestCreateNeed_BRNED01_DatabaseRefusesTheSixthOpenNeed` |
| R-SEC-45 | Lima gagal masuk mengunci 15 menit | Terpenuhi | `api/internal/modules/auth/service/login_test.go` `TestLogin_RSEC45_LockedAccountRefusesEvenTheRightPassword` |

## 7.10 dan 7.11 Validasi dan sanitasi

| Aturan | Ringkas | Status | Bukti |
|---|---|---|---|
| R-SEC-46 | Uji batas tepat, di bawah, di atas | Terpenuhi | `api/internal/modules/org/domain/onboarding_test.go` `TestNewOnboarding_RSEC46_NameLengthBoundaries`; `app/src/lib/features/organization/schema.test.ts` |
| R-SEC-47 | Galat validasi dikumpulkan sekaligus | Terpenuhi | `api/internal/kernel/field/field_test.go` `TestCollect_RSEC47_KeepsEveryProblemInOrder`; `api/cmd/server/products_test.go` `TestCreateProduct_RSEC47_EveryInvalidFieldIsReported` |
| R-SEC-48 | Panjang dihitung per karakter | Terpenuhi | `api/internal/kernel/field/field_test.go` `TestCheckLength_RSEC48_CountsCharactersNotBytes` |
| R-SEC-49 | `lanjut` hanya jalur relatif | Terpenuhi | `app/src/lib/features/auth/links.test.ts` |
| R-SEC-50 | Pangkas, rapatkan spasi, NFC | Terpenuhi | `api/internal/platform/httpserver/decode_test.go` `TestNormalizeLine_RSEC50_TrimsCollapsesAndComposes` |
| R-SEC-51 | Tolak karakter kendali, tanpa lebar, arah | Terpenuhi | `api/internal/platform/httpserver/decode_test.go` `TestNormalizeLine_RSEC51_RejectsControlZeroWidthAndDirectionMarks`; `api/cmd/server/org_test.go` `TestCreateOrganization_RSEC51_HiddenCharactersInANameAreRejected` |
| R-SEC-52 | Email kecil, WhatsApp `62`, slug | Terpenuhi | `api/internal/kernel/contact/whatsapp_test.go` `TestParseWhatsAppNumber_RSEC52_NormalizesTo62`; `api/internal/modules/auth/domain/registration_test.go` `TestNewRegistration_ValidForm_NormalizesEmailAndWhatsApp` |
| R-SEC-53 | Teks polos, tanpa HTML atau Markdown | Terpenuhi | `app/tests/xss.test.ts` (muatan tampil sebagai teks); lint `svelte/no-at-html-tags` |
| R-SEC-54 | Nama berkas dibuang, kunci dari server | Terpenuhi | `api/internal/platform/storage/storage_test.go` `TestPut_RSEC54_KeyIsRandomAndKeepsOnlyTheExtension` |
| R-SEC-55 | Sel CSV berawalan rumus diberi petik | Terpenuhi | `app/src/lib/domain/csv.test.ts` |

## 7.12 Suntikan SQL

| Aturan | Ringkas | Status | Bukti |
|---|---|---|---|
| R-SEC-56 | SQL hanya di `queries` dan `migrations` | Terpenuhi | `gosec` di `api/.golangci.yml`; kode kueri dibangkitkan `sqlc` |
| R-SEC-57 | Hanya `repository` dan `database` mengimpor pgx | Terpenuhi | `depguard` aturan `pgx-boundary` di `api/.golangci.yml` |
| R-SEC-58 | `sort` lewat daftar putih | Terpenuhi | `api/internal/modules/directory/domain/search_test.go` `TestNewStoreSearch_FRDIR06_FiltersAreValidatedTogether` |
| R-SEC-59 | Escape `%`, `_`, `\` untuk `LIKE` | Terpenuhi | `api/internal/modules/product/repository/products_test.go` `TestListProducts_RSEC59_WildcardsInTheSearchAreLiteral`; `api/cmd/server/security_injection_integration_test.go` |
| R-SEC-60 | Peran aplikasi tanpa DDL dan tanpa ubah segel | Belum | Dibuat bersama deploy API |
| R-SEC-61 | Muatan suntikan ke semua parameter cari dan saring | Terpenuhi | `api/cmd/server/security_injection_integration_test.go` `TestSecurity_RSEC61_InjectionPayloadsInSearchAndFiltersAreHarmless` |

## 7.13 dan 7.14 Keluaran dan kepala keamanan

| Aturan | Ringkas | Status | Bukti |
|---|---|---|---|
| R-SEC-62 | `{@html}` dilarang | Terpenuhi | `svelte/no-at-html-tags` di `app/eslint.config.js` |
| R-SEC-63 | Larang `innerHTML`, `eval`, dan sejenisnya | Terpenuhi | `forbiddenSinks`, `no-eval`, `no-new-func`, `@typescript-eslint/no-implied-eval` di `app/eslint.config.js` |
| R-SEC-64 | `safeHref`, tautan WhatsApp, `noopener` | Terpenuhi | `app/src/lib/domain/safe-href.test.ts`; `app/src/lib/domain/whatsapp.test.ts` |
| R-SEC-65 | Isi peta lewat `textContent` | Terpenuhi, tanpa uji khusus | `app/src/lib/components/ui/MapView.svelte`: penanda berupa komponen Svelte, tanpa balon dari data pengguna |
| R-SEC-66 | Data terstruktur dengan `<` di-escape | Belum | Lihat catatan di bawah |
| R-SEC-67 | CSP berbasis hash | Terpenuhi | `app/csp.config.test.ts`; `app/tests/security-headers.test.ts` |
| R-SEC-68 | Tanpa `unsafe-inline` dan `unsafe-eval` di `script-src` | Terpenuhi | `app/csp.config.test.ts`; `app/tests/security-headers.test.ts` |
| R-SEC-69 | Muatan skrip di isian bebas tidak pernah berjalan | Terpenuhi | `app/tests/xss.test.ts` "R-SEC-69 script payloads saved in free text fields never run on the pages that show them" |

## 7.15 dan 7.16 Unggahan, rahasia, rantai pasok

| Aturan | Ringkas | Status | Bukti |
|---|---|---|---|
| R-SEC-70 | Jenis dari isi, maksimal 4096 piksel, kode ulang | Terpenuhi | `api/internal/platform/imaging/imaging_test.go` `TestRecode_RSEC70_DimensionBoundary`, `TestRecode_RSEC70_JPEGIsReencodedWithoutMetadata` |
| R-SEC-71 | Berkas dari asal API, `inline`, `nosniff` | Terpenuhi | `api/cmd/server/files_test.go` `TestUpload_RSEC70_PhotoIsRecodedAndServedPubliclyWithoutLocationData` |
| R-SEC-72 | Berkas privat lewat proxy, tanpa cache | Terpenuhi | `api/cmd/server/files_test.go` `TestUpload_RSEC72_PaymentQRIsPrivateAndOnlyReadableThroughTheAPI` |
| R-SEC-73 | Rahasia dari lingkungan, `.env.example` tanpa nilai | Terpenuhi | `api/cmd/devenv/main_test.go` `TestEnvExamples_RSEC73_ListNamesWithoutValues`; `.gitignore` |
| R-SEC-74 | Server menolak rahasia kosong, pendek, contoh | Terpenuhi | `api/internal/platform/config/config_test.go` `TestLoad_EmptySecret_IsRejected`, `TestLoad_PlaceholderSecret_IsRejected` |
| R-SEC-75 | Pemindai rahasia di kait dan CI | Terpenuhi | `gitleaks` di `lefthook.yml` dan `make check-repo` di CI |
| R-SEC-76 | Pemindai kerentanan, berkas kunci, citra non-root | Belum | `govulncheck` dan `npm audit` sudah berjalan di `make check`; citra Docker dibuat bersama deploy API |
| R-SEC-77 | String koneksi tidak bocor | Terpenuhi | `api/cmd/dbcheck/main_test.go` `TestRun_UnreachableHost_HidesConnectionString`; `api/internal/platform/database/database_test.go` `TestOpen_UnparsableURL_HidesTheURL` |

## Catatan untuk butir yang belum

- **R-SEC-20**: penggantian peran dan penonaktifan anggota adalah bagian dari fitur anggota staf (FR-ORG-09, P1), sehingga belum ada. Penangguhan organisasi saat ini dilakukan pengelola langsung di basis data (`docs/MODERASI.md`); organisasi yang ditangguhkan tidak bisa mengirim Ajuan, tetapi sesi anggotanya baru berakhir saat token aksesnya habis (paling lama 15 menit) atau saat pengelola menghapus barisnya di `sessions`. Pencabutan otomatis ikut dikerjakan bersama P1.2.
- **R-SEC-66**: data terstruktur untuk mesin pencari adalah FR-PUB-10 (P1), sehingga belum ada halaman yang memuatnya.
