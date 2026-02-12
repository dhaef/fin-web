import { buildBarChart, defaultBarColors, netIncomeBarColors } from "widgets";

buildBarChart("yearly-expense-bar", "yearly-expense-counts", defaultBarColors);
buildBarChart("yearly-income-bar", "yearly-income-counts", defaultBarColors);
buildBarChart("yearly-net-income-bar", "yearly-net-counts", netIncomeBarColors);
