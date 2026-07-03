import { donut } from 'widgets';

// Needs / Wants / Savings for the trailing window. Navigation is disabled since
// these wedges aren't categories to filter by.
const breakdownDonut = document.getElementById('breakdown-donut');
const breakdownCountsEl = document.getElementById('breakdown-counts');
if (breakdownDonut && breakdownCountsEl) {
  const counts = JSON.parse(breakdownCountsEl.textContent);

  const { node } = donut(
    counts,
    [
      '#4288b5', // Needs — blue
      '#f0a35e', // Wants — amber
      '#2E865F', // Savings — green
    ],
    true
  );
  breakdownDonut.appendChild(node);
}
