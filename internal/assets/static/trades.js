import { donut } from 'widgets';
import * as d3 from 'd3';

const categoryDonut = document.getElementById('trades-donut');
const categoryCounts = document.getElementById('trades-counts');
if (categoryDonut && categoryCounts) {
  const counts = [];

  for (let i = 0; i < categoryCounts.children.length; i++) {
    const elValue = categoryCounts.children[i].textContent;
    const [id, name, value] = elValue.split(':');
    const numberValue = Number(value);
    counts.push({ id, name, value: numberValue });
  }

  const { node } = donut(
    counts,
    d3
      .quantize((t) => d3.interpolateYlOrRd(t * 0.7 + 0.3), counts.length)
      .reverse(),
    true
  );
  categoryDonut.appendChild(node);
}
