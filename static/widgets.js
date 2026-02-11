import * as d3 from "https://cdn.jsdelivr.net/npm/d3@7/+esm";

function donut(data) {
  const width = 400;
  const height = Math.min(width, 500);
  const radius = Math.min(width, height) / 2;

  const arc = d3
    .arc()
    .innerRadius(radius * 0.67)
    .outerRadius(radius - 1);

  const pie = d3
    .pie()
    .padAngle(1 / radius)
    .sort(null)
    .value((d) => d.value);

  const color = d3
    .scaleOrdinal()
    .domain(data.map((d) => d.name))
    .range(
      d3
        .quantize((t) => d3.interpolateSpectral(t * 0.8 + 0.1), data.length)
        .reverse(),
    );

  const svg = d3
    .create("svg")
    .attr("width", width)
    .attr("height", height)
    .attr("viewBox", [-width / 2, -height / 2, width, height])
    .attr("style", "max-width: 100%; height: auto;");

  const namesAndColors = [];

  svg
    .append("g")
    .selectAll()
    .data(pie(data))
    .join("path")
    .attr("fill", (d) => {
      let c = color(d.data.name);

      if (c === "rgb(0, 0, 0)") {
        c = "rgb(254, 221, 141)";
      }

      namesAndColors.push({
        name: d.data.name,
        color: c,
        value: d.data.value,
      });
      return c;
    })
    .attr("d", arc)
    .on("click", (_, i) => {
      const p = new URLSearchParams(location.search);
      p.set("categories", i.data.name);

      window.location = `${window.location.origin}?${p.toString()}`;
    })
    .append("title")
    .text((d) => `${d.data.name}: ${d.data.value}`);

  const legend = document.createElement("div");
  let total = 0;
  for (const n of namesAndColors) {
    const p = document.createElement("p");
    p.style.backgroundColor = n.color;
    p.style.borderRadius = "5px";
    p.style.padding = ".25rem";
    p.style.margin = ".1rem";
    p.textContent = `${n.name}: ${formatter.format(n.value)}`;

    total += n.value;
    legend.appendChild(p);
  }

  const p = document.createElement("p");
  p.style.borderRadius = "5px";
  p.style.padding = ".25rem";
  p.style.margin = ".1rem";
  p.textContent = `total: ${formatter.format(total)}`;
  legend.appendChild(p);

  return { node: svg.node(), legend };
}

function barChart(data, barColor, labelOffset) {
  // Declare the chart dimensions and margins.
  const width = 1200;
  const height = 500;
  const marginTop = 30;
  const marginRight = 0;
  const marginBottom = 30;
  const marginLeft = 40;

  // Declare the x (horizontal position) scale.
  const x = d3
    .scaleBand()
    .domain(data.map((d) => d.name))
    .range([marginLeft, width - marginRight])
    .padding(0.1);

  // Declare the y (vertical position) scale.
  const y = d3
    .scaleLinear()
    .domain([0, d3.max(data, (d) => d.value)])
    .range([height - marginBottom, marginTop]);

  // Create the SVG container.
  const svg = d3
    .create("svg")
    .attr("width", width)
    .attr("height", height)
    .attr("viewBox", [0, 0, width, height])
    .attr("style", "max-width: 100%; height: auto;");

  // Add a rect for each bar.
  svg
    .append("g")
    // .attr('fill', barColor)
    .selectAll()
    .data(data)
    .join("rect")
    .attr("fill", (d) => (d.sign === "neg" ? "rgb(209, 60, 75)" : barColor))
    .attr("x", (d) => x(d.name))
    .attr("y", (d) => y(d.value))
    .attr("height", (d) => y(0) - y(d.value))
    .attr("width", x.bandwidth())
    .on("click", (_, i) => {
      if (i.name.includes("-")) {
        const [month, year] = i.name.split("-");
        const startDate = `${year}-${month}-01`;

        const lastDayOfMonth = new Date(Number(year), Number(month), 0);
        const endDate = `${year}-${month}-${lastDayOfMonth.getDate()}`;

        const p = new URLSearchParams(location.search);
        p.set("startDate", startDate);
        p.set("endDate", endDate);

        window
          .open(`${window.location.origin}?${p.toString()}`, "_blank")
          .focus();
      }
    });

  // Add the x-axis and label.
  svg
    .append("g")
    .attr("transform", `translate(0,${height - marginBottom})`)
    .call(d3.axisBottom(x).tickSizeOuter(0));

  // Add the y-axis and label, and remove the domain line.
  svg
    .append("g")
    .attr("transform", `translate(${marginLeft},0)`)
    .call(d3.axisLeft(y).tickFormat((y) => y.toFixed()))
    .call((g) => g.select(".domain").remove())
    .call((g) =>
      g
        .append("text")
        .attr("x", -marginLeft)
        .attr("y", 10)
        .attr("fill", "currentColor")
        .attr("text-anchor", "start")
        .text("Amount ($)"),
    );

  svg
    .selectAll("text.bar")
    .data(data)
    .enter()
    .append("text")
    .attr("class", "bar")
    .attr("text-anchor", "middle")
    .attr("x", function (d) {
      return x(d.name) + labelOffset;
    })
    .attr("y", function (d) {
      return y(d.value) - 5;
    })
    .attr("style", "font-size: 10px;")
    .text(function (d) {
      if (d.subValue) {
        return `${formatter.format(d.sign === "neg" ? -d.value : d.value)} (${d.subValue})`;
      }
      return formatter.format(d.sign === "neg" ? -d.value : d.value);
    });

  // Return the SVG element.
  return { node: svg.node() };
}

function lineChart(data) {
  // --- PRE-PROCESSING ---
  // 1. Ensure dates are real Date objects and sorted (Required for bisector to work)
  data.forEach((d) => {
    d.date = d.date instanceof Date ? d.date : new Date(d.date);
  });
  data.sort((a, b) => a.date - b.date);

  const width = 1200;
  const height = 500;
  const marginTop = 20;
  const marginRight = 30;
  const marginBottom = 30;
  const marginLeft = 80;

  const x = d3.scaleUtc(
    d3.extent(data, (d) => d.date),
    [marginLeft, width - marginRight],
  );

  const y = d3
    .scaleLinear(
      d3.extent(data, (d) => d.value),
      [height - marginBottom, marginTop],
    )
    .nice();

  const line = d3
    .line()
    .x((d) => x(d.date))
    .y((d) => y(d.value))
    .curve(d3.curveMonotoneX);

  const area = d3
    .area()
    .x((d) => x(d.date))
    .y0(y(0))
    .y1((d) => y(d.value))
    .curve(d3.curveMonotoneX);

  const svg = d3
    .create("svg")
    .attr("width", width)
    .attr("height", height)
    .attr("viewBox", [0, 0, width, height])
    .attr(
      "style",
      "max-width: 100%; height: auto; font-family: sans-serif; overflow: visible;",
    );

  // Axes
  svg
    .append("g")
    .attr("transform", `translate(0,${height - marginBottom})`)
    .call(
      d3
        .axisBottom(x)
        .ticks(width / 100)
        .tickFormat(d3.timeFormat("%b %Y"))
        .tickSizeOuter(0),
    );

  svg
    .append("g")
    .attr("transform", `translate(${marginLeft},0)`)
    .call(
      d3
        .axisLeft(y)
        .ticks(height / 50)
        .tickFormat((d) => d3.format("$.2s")(d).replace("G", "B")),
    )
    .call((g) => g.select(".domain").remove())
    .call((g) =>
      g
        .selectAll(".tick line")
        .clone()
        .attr("x2", width - marginLeft - marginRight)
        .attr("stroke-opacity", 0.1),
    );

  // Zero Line
  svg
    .append("line")
    .attr("x1", marginLeft)
    .attr("x2", width - marginRight)
    .attr("y1", y(0))
    .attr("y2", y(0))
    .attr("stroke", "#333")
    .attr("stroke-dasharray", "4,4")
    .attr("opacity", 0.4);

  // Paths
  svg
    .append("path")
    .datum(data)
    .attr("fill", "rgba(66, 136, 181, 0.1)")
    .attr("d", area);
  svg
    .append("path")
    .datum(data)
    .attr("fill", "none")
    .attr("stroke", "rgb(66, 136, 181)")
    .attr("stroke-width", 2.5)
    .attr("d", line);

  // --- TOOLTIP LOGIC ---
  const tooltip = svg.append("g").style("display", "none");

  // Vertical tracker line
  tooltip
    .append("line")
    .attr("stroke", "#999")
    .attr("stroke-width", 1)
    .attr("stroke-dasharray", "3,3")
    .attr("y1", marginTop)
    .attr("y2", height - marginBottom);

  // Circle marker
  tooltip
    .append("circle")
    .attr("r", 6)
    .attr("fill", "rgb(66, 136, 181)")
    .attr("stroke", "white")
    .attr("stroke-width", 2);

  // Tooltip Label Group
  const label = tooltip.append("g").attr("transform", "translate(0, -35)"); // Moved up to fit two lines

  // Background for the label
  label
    .append("rect")
    .attr("fill", "white")
    .attr("fill-opacity", 0.9)
    .attr("stroke", "#ccc")
    .attr("x", -60)
    .attr("width", 120)
    .attr("height", 45) // Taller for two lines
    .attr("rx", 4);

  // Date text (Top line)
  label
    .append("text")
    .attr("class", "tooltip-date")
    .attr("text-anchor", "middle")
    .attr("dy", 18)
    .attr("font-size", "12px")
    .attr("fill", "#666");

  // Value text (Bottom line)
  label
    .append("text")
    .attr("class", "tooltip-value")
    .attr("text-anchor", "middle")
    .attr("dy", 36)
    .attr("font-size", "14px")
    .attr("font-weight", "bold")
    .attr("fill", "#333");

  const bisectDate = d3.bisector((d) => d.date).left;
  const formatDate = d3.timeFormat("%b %Y"); // e.g., "Jan 2024"

  svg
    .append("rect")
    .attr("width", width)
    .attr("height", height)
    .attr("fill", "none")
    .attr("pointer-events", "all")
    .on("mouseover", () => tooltip.style("display", null))
    .on("mouseout", () => tooltip.style("display", "none"))
    .on("mousemove", (event) => {
      const x0 = x.invert(d3.pointer(event)[0]);
      let i = bisectDate(data, x0, 1);

      if (i >= data.length) i = data.length - 1;
      const d0 = data[i - 1];
      const d1 = data[i];
      const d = x0 - d0.date > d1.date - x0 ? d1 : d0;

      tooltip.attr("transform", `translate(${x(d.date)},0)`);
      tooltip.select("circle").attr("cy", y(d.value));

      // Update the two lines of text
      tooltip.select(".tooltip-date").text(formatDate(d.date));
      tooltip.select(".tooltip-value").text(d3.format("$,.0f")(d.value));
    });

  return { node: svg.node() };
}

const categoryDonut = document.getElementById("category-donut");
const categoryCounts = document.getElementById("category-counts");
if (categoryDonut && categoryCounts) {
  const counts = [];

  for (let i = 0; i < categoryCounts.children.length; i++) {
    const elValue = categoryCounts.children[i].textContent;
    const [name, value] = elValue.split(":");
    const numberValue = Number(value);
    counts.push({ name, value: numberValue });
  }

  const { node, legend } = donut(counts);
  categoryDonut.appendChild(legend);
  categoryDonut.appendChild(node);
}

const categoryIncomeDonut = document.getElementById("category-income-donut");
const categoryIncomeCounts = document.getElementById("category-income-counts");
if (categoryIncomeDonut && categoryIncomeCounts) {
  const counts = [];

  for (let i = 0; i < categoryIncomeCounts.children.length; i++) {
    const elValue = categoryIncomeCounts.children[i].textContent;
    const [name, value] = elValue.split(":");
    const numberValue = Number(value);
    counts.push({ name, value: Math.abs(numberValue) });
  }

  const { node, legend } = donut(counts);
  categoryIncomeDonut.appendChild(legend);
  categoryIncomeDonut.appendChild(node);
}

function buildBarChart(barId, countsId, color, labelOffset) {
  const bar = document.getElementById(barId);
  const countElements = document.getElementById(countsId);
  if (bar && countElements) {
    const counts = [];
    const children = Array.from(countElements.children).sort((a, b) => {
      const [aName] = a.textContent.split(":");
      const [bName] = b.textContent.split(":");
      const aMonthYear = aName.split("-");
      const bMonthYear = bName.split("-");
      const aDate = new Date();
      aDate.setMonth(Number(aMonthYear[0]) - 1);
      aDate.setYear(Number(aMonthYear[1]) - 1);
      const bDate = new Date();
      bDate.setMonth(Number(bMonthYear[0]) - 1);
      bDate.setYear(Number(bMonthYear[1]) - 1);

      return aDate.getTime() - bDate.getTime();
    });

    for (let i = 0; i < children.length; i++) {
      const elValue = children[i].textContent;
      const [name, value, subValue] = elValue.split(":");
      const numberValue = Number(value);
      counts.push({
        name,
        value: Math.abs(numberValue),
        subValue,
        sign: numberValue > 0 ? "neg" : undefined,
      });
    }

    const { node } = barChart(counts, color, labelOffset);
    bar.appendChild(node);
  }
}

// making this specific to net-worth but could make it more generic later
function buildLineChart(chartId) {
  const chart = document.getElementById(chartId);
  if (chart) {
    const netWorthItems = document.querySelectorAll(".net-worth");
    const data = [];
    const parseDate = d3.utcParse("%Y-%m-%d");
    for (const nwi of Array.from(netWorthItems)) {
      const value = Number(nwi.getAttribute("data-value"));
      const date = nwi.getAttribute("data-date");

      data.push({
        date: parseDate(date),
        value,
      });
    }

    const { node } = lineChart(data);
    chart.appendChild(node);
  }
}
buildLineChart("net-worth-line-chart");

buildBarChart(
  "previous-year-bar",
  "previous-year-counts",
  "rgb(254, 221, 141)",
  44,
);
buildBarChart(
  "previous-year-income-bar",
  "previous-year-income-counts",
  "rgb(114, 195, 167)",
  44,
);
buildBarChart("net-income-bar", "net-income-counts", "rgb(66, 136, 181)", 44);
buildBarChart(
  "yearly-expense-bar",
  "yearly-expense-counts",
  "rgb(254, 221, 141)",
  103,
);
buildBarChart(
  "yearly-income-bar",
  "yearly-income-counts",
  "rgb(114, 195, 167)",
  103,
);
buildBarChart(
  "yearly-net-income-bar",
  "yearly-net-counts",
  "rgb(66, 136, 181)",
  103,
);
