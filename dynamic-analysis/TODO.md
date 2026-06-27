# TODO - Network sim enhanced evaluation

## Step 1: Diagnose Docker INetSim network creation
- [x] Inspect Docker networks on the machine
- [x] Find that `service-simulation-module_simulation_network` uses subnet `172.20.0.0/24`
- [ ] Update `docker-compose.network-sim.yml` to use a non-overlapping subnet for `pack-a-mal-network`

## Step 2: Re-run Bước 3 end-to-end
- [ ] `docker-compose -f docker-compose.network-sim.yml up -d`
- [ ] Run 3 sample suites and collect logs for baseline/enhanced
- [ ] Compute stages+anchors using `logs/count_results.py`
- [ ] Paste 4 lines of output into LaTeX update

