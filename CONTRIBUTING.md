# Contributing to musegpt

First off, thank you for considering contributing to **musegpt**! Your contributions help improve the project and are greatly appreciated.

This document outlines the guidelines for contributing to the **musegpt** project. Following these guidelines ensures a smooth collaboration between all contributors.

## Table of Contents

- [How Can I Contribute?](#how-can-i-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Features](#suggesting-features)
  - [Contributing Code](#contributing-code)
  - [Improving Documentation](#improving-documentation)
- [Getting Started](#getting-started)
  - [Environment Setup](#environment-setup)
  - [Building the Project](#building-the-project)
  - [Running Tests](#running-tests)
- [Coding Guidelines](#coding-guidelines)
  - [Style Guide](#style-guide)
  - [Commit Messages](#commit-messages)
  - [Branch Management](#branch-management)
- [Pull Request Process](#pull-request-process)
  - [Before Submitting](#before-submitting)
  - [Submitting a Pull Request](#submitting-a-pull-request)
- [Code of Conduct](#code-of-conduct)
- [License](#license)

## How Can I Contribute?

There are several ways you can contribute to **musegpt**:

### Reporting Bugs

If you encounter any issues or bugs, please report them using the [issue tracker](https://github.com/greynewell/musegpt/issues).

**When reporting a bug, please include:**

- A clear and descriptive title.
- Steps to reproduce the issue.
- Expected and actual results.
- Any relevant logs or screenshots.

### Suggesting Features

We welcome new ideas to enhance **musegpt**. Feel free to submit feature requests via the [issue tracker](https://github.com/greynewell/musegpt/issues).

**When suggesting a feature, please include:**

- A clear explanation of the feature and its benefits.
- Any additional context or references.

### Contributing Code

You can contribute by fixing bugs, implementing new features, or improving existing code.

### Improving Documentation

Help us keep the documentation up-to-date and comprehensive. You can:

- Fix typos or grammatical errors.
- Add missing information.
- Enhance explanations for clarity.

## Getting Started

### Environment Setup

To set up your development environment:

1. **Clone the repository:**

   ```bash
   git clone https://github.com/greynewell/musegpt.git
   cd musegpt
   ```

2. **Install dependencies:**

   Ensure you have the following installed:

   - **C++17 compatible compiler** (e.g., GCC 7+, Clang 5+, MSVC 2017+)
   - **[CMake](https://cmake.org/)** 3.15 or later
   - **[JUCE](https://juce.com/)** (Audio application framework)
   - **[llama.cpp](https://github.com/ggerganov/llama.cpp)** (LLM inference library)

### Building the Project

You can build the project using the provided scripts:

- **Debug build:**

  ```bash
  ./scripts/build/debug.sh
  ```

- **Release build:**

  ```bash
  ./scripts/build/release.sh
  ```

### Running Tests

(Currently, automated tests are not implemented. Contributions to add testing capabilities are welcome!)

## Coding Guidelines

### Style Guide

Please adhere to the following coding standards:

- **Follow the [JUCE Coding Standards](https://juce.com/discover/stories/coding-standards).**
- **Use consistent indentation and formatting.**
- **Write clear and maintainable code with comments where necessary.**
- **Use meaningful variable and function names.**

### Commit Messages

- Use the *present tense* (e.g., "Add feature" not "Added feature").
- Be concise but descriptive.
- Reference relevant issues or pull requests when applicable (e.g., "Fix memory leak in audio processing module. Closes #42").

### Branch Management

- **Main branch:** `main` (stable codebase)
- **Feature branches:** Create a new branch for each feature or bug fix (e.g., `feature/midi-integration`).
- **Keep branches up-to-date** by regularly merging or rebasing with `main`.

## Pull Request Process

### Before Submitting

- Ensure your code compiles and runs without errors.
- Test your changes thoroughly.
- Update the documentation if your changes affect it.
- Check for code style compliance.

### Submitting a Pull Request

1. **Fork the repository** and create your feature branch from `main`:

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Commit your changes** with clear messages.

3. **Push to your forked repository**:

   ```bash
   git push origin feature/your-feature-name
   ```

4. **Open a pull request** to the `main` branch of the original repository.

5. **Provide a detailed description** of your changes and the problem they solve.

### Code Review

- Your pull request will be reviewed by project maintainers.
- Be open to feedback and willing to make necessary changes.
- Engage in discussions to clarify any questions.

## Code of Conduct

By participating in this project, you agree to abide by the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). Please read it to understand the expectations for all contributors.

## License

By contributing to **musegpt**, you agree that your contributions will be licensed under the [AGPL v3 License](LICENSE).

---

Thank you for helping to make **musegpt** better! Your contributions are invaluable to the growth and success of the project.