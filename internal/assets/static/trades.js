import { donut } from 'widgets';
import * as d3 from 'd3';

const tradesDonut = document.getElementById('trades-donut');
const tradesCountsEl = document.getElementById('trades-counts');
if (tradesDonut && tradesCountsEl) {
  const counts = JSON.parse(tradesCountsEl.textContent);

  const { node } = donut(
    counts,
    d3
      .quantize((t) => d3.interpolateYlOrRd(t * 0.7 + 0.3), counts.length)
      .reverse(),
    true
  );
  tradesDonut.appendChild(node);
}
