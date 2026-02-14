import { buildBarChart, defaultBarColors, netIncomeBarColors } from "widgets";

buildBarChart("yearly-expense-bar", "yearly-expense-counts", defaultBarColors);
buildBarChart("yearly-income-bar", "yearly-income-counts", defaultBarColors);
buildBarChart("yearly-net-income-bar", "yearly-net-counts", netIncomeBarColors);

document.getElementById("chart-select").addEventListener("change", (e) => {
  const el = e.target;

  for (const child of el.children) {
    const chart = document.getElementById(child.value);

    if (child.value === el.value) {
      chart.classList.remove("hide");
    } else {
      chart.classList.add("hide");
    }
  }
});
