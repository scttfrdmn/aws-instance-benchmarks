---
name: Benchmark Data Contribution
about: Submit new benchmark results for the community database
title: '[DATA] New benchmark results for '
labels: data, contribution
assignees: ''

---

**Instance Types**
List the AWS instance types for which you're contributing benchmark data:
- [ ] Instance type 1 (e.g., m7i.large)
- [ ] Instance type 2 (e.g., c7g.xlarge)

**Benchmark Suites**
Which benchmarks did you run?
- [ ] STREAM (memory bandwidth)
- [ ] HPL/LINPACK (CPU performance)
- [ ] Custom benchmark (please describe)

**Regions**
Which AWS regions were used for testing?
- [ ] us-east-1
- [ ] us-west-2
- [ ] eu-west-1
- [ ] Other: ___________

**Data Quality Checklist**
- [ ] Multiple runs completed (minimum 3 iterations)
- [ ] Results are consistent across runs
- [ ] Schema validation passes
- [ ] No anomalous results that can't be explained
- [ ] Proper statistical analysis included

**Methodology**
Briefly describe your benchmark methodology:
- Container used (if not standard)
- Compilation flags or optimizations
- Any special configuration

**Files**
Please attach or link to:
- [ ] Benchmark result JSON files
- [ ] Validation output
- [ ] Any additional metadata

**Validation**
Have you validated these results?
- [ ] Compared against known baselines
- [ ] Verified pricing calculations
- [ ] Checked for outliers
- [ ] Schema validation passes

**Additional Context**
Any additional information about the benchmark run:
- Special conditions or configurations
- Known issues or limitations
- Comparison with existing data

**Contribution Agreement**
- [ ] I agree to contribute this data under the project's open license
- [ ] I confirm this data was collected ethically and legally
- [ ] I understand this data will be publicly available