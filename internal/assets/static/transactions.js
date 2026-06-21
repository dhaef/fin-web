import {
  buildBarChart,
  defaultBarColors,
  netIncomeBarColors,
  donut,
} from 'widgets';
import * as d3 from 'd3';

buildBarChart('previous-year-bar', 'previous-year-counts', defaultBarColors);
buildBarChart(
  'previous-year-income-bar',
  'previous-year-income-counts',
  defaultBarColors
);
buildBarChart('net-income-bar', 'net-income-counts', netIncomeBarColors);

document.getElementById('chart-select').addEventListener('change', (e) => {
  const el = e.target;

  for (const child of el.children) {
    const chart = document.getElementById(child.value);

    if (child.value === el.value) {
      chart.classList.remove('hide');
    } else {
      chart.classList.add('hide');
    }
  }
});

const categoryDonut = document.getElementById('category-donut');
const categoryCountsEl = document.getElementById('category-counts');
if (categoryDonut && categoryCountsEl) {
  const counts = JSON.parse(categoryCountsEl.textContent);

  const { node } = donut(
    counts,
    d3
      .quantize((t) => d3.interpolateYlOrRd(t * 0.7 + 0.3), counts.length)
      .reverse()
  );
  categoryDonut.appendChild(node);
}

const categoryIncomeDonut = document.getElementById('category-income-donut');
const categoryIncomeCountsEl = document.getElementById('category-income-counts');
if (categoryIncomeDonut && categoryIncomeCountsEl) {
  const counts = JSON.parse(categoryIncomeCountsEl.textContent).map((d) => ({
    ...d,
    value: Math.abs(d.value),
  }));

  const { node } = donut(counts, [
    '#2E865F', // Deep Emerald (Main Income)
    '#66b3a1', // Soft Teal
    '#a5d6a7', // Fresh Mint
    '#26a69a', // Darker Teal
    '#81c784', // Sage Green
    '#b2dfdb', // Light Aqua
  ]);
  categoryIncomeDonut.appendChild(node);
}
