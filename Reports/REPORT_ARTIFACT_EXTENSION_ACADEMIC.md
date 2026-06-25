# Chapter 5. Evaluation

This chapter presents the evaluation of the implemented system. The objective is to assess whether the proposed network simulation and artifact emulation approach improves sandbox realism and supports more comprehensive behavioral observation during dynamic analysis. Rather than focusing solely on binary success or failure, the evaluation emphasizes behavior activation, trace quality, and reproducibility.

The evaluation addresses three key questions. First, does the proposed system enable execution paths that would otherwise remain inactive because of missing external dependencies or an unrealistic environment? Second, does the presence of simulated artifacts improve the coherence and interpretability of analysis traces? Third, does the system produce stable and reproducible behavior across repeated executions under the same configuration? Together, these questions assess not only functional correctness but also the practical usefulness of the system for malware analysis workflows.

## 5.1 Evaluation Setup

The evaluation is organized around a controlled comparison between baseline execution and the enhanced analysis environment. The baseline represents a minimal sandbox configuration without network simulation or artifact injection. The enhanced environment combines adaptive network simulation with persistent artifact emulation in order to approximate a more realistic host.

The system was evaluated in an isolated execution environment with the following configurations:

- **Baseline setup:** no network simulation and no artifact injection.
- **Network simulation only:** DNS and HTTP traffic are redirected through INetSim, but no persistent host artifacts are injected.
- **Network simulation plus artifact emulation:** network simulation is enabled and the host environment is enriched with realistic persistent artifacts.

The comparison is designed to measure how much additional execution behavior becomes observable once the environment is made more realistic.

## 5.2 Sample Description

The evaluation uses **three representative test samples**, all created in-house and stored under `dynamic-analysis/sample_packages/malicious_network_package/`:

- [test_network.py](https://github.com/DangTheNhan/EIU-Chat-Zone/blob/main/dynamic-analysis/sample_packages/malicious_network_package/test_network.py)
- [test_with_inetsim.py](https://github.com/DangTheNhan/EIU-Chat-Zone/blob/main/dynamic-analysis/sample_packages/malicious_network_package/test_with_inetsim.py)
- [test_full_mode.py](https://github.com/DangTheNhan/EIU-Chat-Zone/blob/main/dynamic-analysis/sample_packages/malicious_network_package/test_full_mode.py)

These samples are **synthetic test scripts**, not real malware samples collected from the wild. They were intentionally authored to emulate common malicious network behaviors in a controlled and reproducible way, so that the impact of network simulation and artifact emulation could be observed under consistent conditions.

The samples were selected to cover three distinct execution scenarios:

1. Direct execution without simulation.
2. Execution with INetSim redirection.
3. Execution with the stricter Full Mode interception policy.

The baseline for comparison is the **non-simulated environment**, meaning that no network simulation and no artifact injection are enabled. Under this baseline, the samples are expected to fail early because DNS resolution and HTTP requests are not intercepted, and no persistent host artifacts are available.

For the enhanced configurations, the samples are executed with either:

- **Network simulation only**, where traffic is redirected through INetSim but the host remains minimal.
- **Network simulation plus artifact emulation**, where INetSim is combined with the artifact extension so that the environment also contains realistic persistent artifacts.

This sample design allows the evaluation to answer three practical questions:

- How many execution paths become active when simulation is enabled?
- Does the resulting trace become more coherent and easier to interpret?
- Are repeated runs stable under the same configuration?

In short, the sample set is intentionally small, controlled, and reproducible. Its purpose is to provide a reliable benchmark for evaluating sandbox realism rather than to represent a large malware corpus.

## 5.3 Experimental Results

The results indicate clear differences between the baseline and enhanced configurations. In baseline runs without simulation, execution frequently terminates early because network requests fail or environmental conditions are not met. As a result, traces often contain repeated failure events with limited contextual value.

When network simulation is enabled, the samples are more likely to proceed beyond initial dependency checks. Requests that would otherwise fail are resolved through simulated responses, allowing subsequent execution stages to become observable. This leads to deeper behavior activation and a more complete representation of the intended execution logic.

Artifact emulation further improves analysis quality by reducing indicators commonly associated with artificial environments. The presence of realistic system artifacts results in fewer early termination paths and more consistent execution flow across runs. Traces generated under this configuration exhibit clearer relationships between network activity, system changes, and execution transitions.

**Table 5.1: Execution stage visibility under different configurations**

| Configuration | Initial Stage | Mid Stage | Late Stage |
|---|---|---|---|
| No simulation | Usually visible | Often interrupted | Rarely visible |
| Network simulation only | Visible | Frequently visible | Sometimes visible |
| Network simulation + artifact emulation | Visible | Usually visible | Frequently visible |

As shown in Table 5.1, the combination of network simulation and artifact emulation enables deeper execution stages to become observable more consistently.

## 5.4 Discussion

A comparative analysis highlights the advantages of the proposed approach. Without simulation, trace quality is often fragmented, making it difficult to establish causal links between events. With simulation enabled, request-response interactions serve as anchor points that connect subsequent actions, improving narrative coherence.

Reproducibility is also enhanced, as minor timing differences or network failures in baseline conditions can cause significant divergence between runs. Under controlled simulation policies, execution paths become more stable, allowing meaningful comparisons between different samples or configurations.

From an analyst's perspective, the enhanced setup also reduces interpretation effort because traces contain more complete behavior chains and require fewer assumptions when documenting findings. **Table 5.2** summarizes the observed differences across the main analysis quality dimensions.

**Table 5.2: Comparison of analysis quality dimensions**

| Dimension | Baseline | Enhanced System |
|---|---|---|
| Behavior activation | Low to moderate | Moderate to high |
| Trace coherence | Fragmented | Structured |
| Run-to-run stability | Variable | Stable |
| Interpretation effort | High | Lower |

The enhanced system consistently produces more structured traces with clearer causal relationships between events. Reduced run-to-run variance also improves reproducibility, which is particularly important for systematic analysis and reporting.

These observations show that sandbox enhancement should not be measured solely by whether malware executes, but also by the quality and interpretability of the resulting traces. Simulation profiles increase behavior activation, but they must be balanced against realism to avoid triggering implausible execution paths.

The results therefore suggest that adaptive and selective simulation provides a practical middle ground: by simulating only unreachable dependencies and focusing on high-value artifacts, the system improves analysis depth without introducing excessive artificiality. This balance is especially important in academic and research contexts, where methodological transparency and reproducibility are essential.

## 5.5 Evaluation Summary

In summary, the evaluation shows that the proposed system improves dynamic analysis outcomes by enabling deeper behavior activation, producing more coherent traces, and enhancing reproducibility. These improvements directly support more reliable and interpretable malware analysis, validating the design choices presented in earlier chapters.

## 5.6 Conclusion

The system was evaluated against a baseline environment that lacked both network simulation and artifact emulation. Compared with that baseline, the enhanced environment produced more complete execution traces and more consistent behavior across repeated runs.

The use of three synthetic test samples made the evaluation controlled and reproducible. The samples were deliberately created to represent different network-dependent execution paths, allowing the study to observe how simulation and artifact injection influence execution visibility.

Overall, the results confirm that the proposed approach is effective for improving sandbox realism and for making dynamic analysis traces more informative, structured, and practical for human interpretation.
