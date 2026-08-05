# Changelog

All notable changes to the Facturino Go SDK are documented here. This project
adheres to [Semantic Versioning](https://semver.org/) and
[Keep a Changelog](https://keepachangelog.com/).

## [1.1.0] - 2026-08-05

### Added
- Document `paypal` as an accepted payment method (`PaymentParams.Method`). The
  field is a plain string, so the value already passes through; the API may add
  further payment method values over time — tolerate unknown values.

## [1.0.0] - 2026-07-25 — Initial release
