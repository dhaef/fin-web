import * as d3 from "d3";

export function donut(data, colorRange) {
  // 1. Setup dimensions and data calculations
  const width = 400;
  const height = 400;
  const radius = Math.min(width, height) / 2;
  const totalAmount = d3.sum(data, (d) => d.value);

  // 2. Generators
  const arc = d3
    .arc()
    .innerRadius(radius * 0.65)
    .outerRadius(radius - 10);

  const arcHover = d3
    .arc()
    .innerRadius(radius * 0.65)
    .outerRadius(radius);

  const pie = d3
    .pie()
    .padAngle(2 / radius)
    .sort(null)
    .value((d) => d.value);

  // 3. Color Scale
  const color = d3.scaleOrdinal().domain(data.map((d) => d.name));

  if (colorRange) {
    color.range(colorRange);
  } else {
    color.range(
      d3
        .quantize((t) => d3.interpolateSpectral(t * 0.8 + 0.1), data.length)
        .reverse(),
    );
  }

  // 4. Create SVG
  const svg = d3
    .create("svg")
    .attr("width", width)
    .attr("height", height)
    .attr("viewBox", [-width / 2, -height / 2, width, height])
    .attr("style", "max-width: 100%; height: auto; font-family: sans-serif;");

  // 5. Center Labels (Total)
  const centerText = svg.append("g").attr("text-anchor", "middle");

  const totalLabel = centerText
    .append("text")
    .attr("dy", "-0.5em")
    .attr("fill", "#666")
    .attr("font-size", "14px")
    .text("Total");

  const totalValue = centerText
    .append("text")
    .attr("dy", "1em")
    .attr("font-size", "22px")
    .attr("font-weight", "bold")
    .text(formatter.format(totalAmount));

  // 6. Draw Arcs with Interactions
  svg
    .append("g")
    .selectAll()
    .data(pie(data))
    .join("path")
    .attr("fill", (d) => color(d.data.name))
    .attr("d", arc)
    .attr("cursor", "pointer")
    .style("transition", "all 0.2s ease")
    .on("mouseenter", function (_event, d) {
      d3.select(this).attr("d", arcHover);
      totalLabel.text(d.data.name);
      totalValue.text(formatter.format(d.data.value));
    })
    .on("mouseleave", function () {
      d3.select(this).attr("d", arc);
      totalLabel.text("Total");
      totalValue.text(formatter.format(totalAmount));
    })
    .on("click", (_event, d) => {
      const p = new URLSearchParams(location.search);
      p.set("categories", d.data.id);
      window.location = `${window.location.origin}?${p.toString()}`;
    });

  return { node: svg.node() };
}

function barChart(data, id, colors) {
  // Dimensions & Scales
  const width = 1200;
  const height = 500;
  const marginTop = 50;
  const marginRight = 20;
  const marginBottom = 40;
  const marginLeft = 60;

  const x = d3
    .scaleBand()
    .domain(data.map((d) => d.name))
    .range([marginLeft, width - marginRight])
    .padding(0.2);

  const y = d3
    .scaleLinear()
    .domain([0, d3.max(data, (d) => d.value)])
    .range([height - marginBottom, marginTop])
    .nice();

  const svg = d3
    .create("svg")
    .attr("width", width)
    .attr("height", height)
    .attr("viewBox", [0, 0, width, height])
    .attr("style", "max-width: 100%; height: auto; font-family: sans-serif;")
    .attr("id", `${id}-chart`);

  // 1. Grid Lines (Bottom Layer)
  svg
    .append("g")
    .attr("transform", `translate(${marginLeft},0)`)
    .call(d3.axisLeft(y).ticks(5).tickFormat(d3.format("$.2s")))
    .call((g) => g.select(".domain").remove())
    .call((g) =>
      g
        .selectAll(".tick line")
        .clone()
        .attr("x2", width - marginLeft - marginRight)
        .attr("stroke-opacity", 0.1),
    );

  // 2. Value Labels (Hidden by default)
  svg
    .append("g")
    .selectAll("text")
    .data(data)
    .join("text")
    .attr("id", (_d, i) => `${id}-label-${i}`) // Using ID for faster selection
    .attr("x", (d) => x(d.name) + x.bandwidth() / 2)
    .attr("y", (d) => y(d.value) - 12)
    .attr("text-anchor", "middle")
    .attr("fill", "#333")
    .style("font-size", "12px")
    .style("font-weight", "bold")
    .style("opacity", 0)
    .style("pointer-events", "none")
    .style("transition", "opacity 0.2s ease, transform 0.2s ease") // Smooth CSS transition
    .text((d) => formatter.format(d.sign === "neg" ? -d.value : d.value));

  // 3. Bars
  svg
    .append("g")
    .selectAll("rect")
    .data(data)
    .join("rect")
    .attr("x", (d) => x(d.name))
    .attr("y", (d) => y(d.value))
    .attr("height", (d) => y(0) - y(d.value))
    .attr("width", x.bandwidth())
    .attr("fill", (d) => (d.sign === "neg" ? colors.negColor : colors.posColor))
    .attr("rx", 4)
    .attr("cursor", "pointer")
    // Use standard CSS transitions for smoother, non-flickering animations
    .style("transition", "fill-opacity 0.2s, transform 0.2s ease")
    .style("transform-origin", "bottom")
    .on("mouseenter", function (_event, d) {
      const index = data.indexOf(d);

      // 1. Highlight Bar using CSS Transform (prevents height math bugs)
      d3.select(this)
        .style("fill-opacity", 0.8)
        .style("transform", "scaleY(1.02)"); // Subtle 2% growth upwards

      // 2. Show Label
      d3.select(`#${id}-label-${index}`)
        .style("opacity", 1)
        .style("transform", "translateY(-5px)"); // Float the text up slightly
    })
    .on("mouseleave", function (_event, d) {
      const index = data.indexOf(d);

      // 1. Reset Bar
      d3.select(this).style("fill-opacity", 1).style("transform", "scaleY(1)");

      // 2. Hide Label
      d3.select(`#${id}-label-${index}`)
        .style("opacity", 0)
        .style("transform", "translateY(0)");
    })
    .on("click", function (_event, d) {
      if (
        !location.pathname.startsWith("/transactions") &&
        !location.pathname.startsWith("/")
      )
        return;

      const [month, year] = d.name.split("-").map(Number);

      const startDate = new Date(year, month - 1, 1);
      // By passing 'month' (which is actually the next month index) and '0' as the day
      const endDate = new Date(year, month, 0);

      const p = new URLSearchParams(location.search);
      p.set("startDate", startDate.toISOString().split("T")[0]);
      p.set("endDate", endDate.toISOString().split("T")[0]);
      window.location = `${window.location.origin}?${p.toString()}`;
    });

  // 4. X-Axis (Top Layer)
  svg
    .append("g")
    .attr("transform", `translate(0,${height - marginBottom})`)
    .call(d3.axisBottom(x).tickSizeOuter(0));

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

export function buildBarChart(barId, countsId, colors) {
  if (document.getElementById(`${barId}-chart`)) {
    return;
  }

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

    const { node } = barChart(counts, barId, colors);
    bar.appendChild(node);
  }
}

// making this specific to net-worth but could make it more generic later
export function buildLineChart(chartId) {
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

export const defaultBarColors = {
  posColor: "#2E865F",
  negColor: "#D13C4B",
};

export const netIncomeBarColors = {
  posColor: "#4288b5",
  negColor: "#D13C4B",
};
