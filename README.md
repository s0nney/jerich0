jerich0
=====================

# HIGHLY EXPERIMENTAL


**A streamlined fork of github.com/dchest/captcha without audio support**

This package implements generation and verification of image CAPTCHAs, forked from dchest/captcha with audio capabilities removed. It provides visual challenges with letters and numbers designed to be difficult for OCR systems to solve.

Key differences from the original:
- Audio CAPTCHA support completely removed
- Supports alphanumeric characters (A-Z, 0-9) instead of just digits
- Simplified codebase focused only on image CAPTCHAs
- Case-sensitive verification 

An image representation is a PNG-encoded image with the solution printed in a distorted way that makes it hard for computers to solve using OCR.

This package doesn't require external files or libraries; it's self-contained.

### Security Note
While this provides basic bot protection, advanced OCR systems may eventually solve these CAPTCHAs. Consider this a first line of defense rather than absolute protection.

## Installation
```bash
go get github.com/s0nney/jerich0
