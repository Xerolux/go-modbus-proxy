<template>
  <div class="h-full w-full p-3">
    <div v-if="loading" class="flex items-center justify-center h-full">
      <i class="pi pi-spin pi-spinner text-2xl"></i>
    </div>
    <div v-else ref="chartContainer" class="h-full w-full"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue';

const props = defineProps({
  data: {
    type: Array,
    required: true
  },
  title: {
    type: String,
    default: 'Chart'
  },
  type: {
    type: String,
    default: 'line' // line, bar
  },
  dataKey: {
    type: String,
    default: 'value'
  },
  labelKey: {
    type: String,
    default: 'label'
  },
  color: {
    type: String,
    default: '#3b82f6'
  }
});

const loading = ref(true);
const chartContainer = ref(null);
let chart = null;

const initChart = () => {
  if (!chartContainer.value) return;

  // Simple canvas-based chart (no external dependencies)
  const canvas = document.createElement('canvas');
  canvas.style.width = '100%';
  canvas.style.height = '100%';
  chartContainer.value.innerHTML = '';
  chartContainer.value.appendChild(canvas);

  const ctx = canvas.getContext('2d');
  const width = chartContainer.value.clientWidth;
  const height = chartContainer.value.clientHeight;

  canvas.width = width * window.devicePixelRatio;
  canvas.height = height * window.devicePixelRatio;
  ctx.scale(window.devicePixelRatio, window.devicePixelRatio);

  if (props.data.length === 0) {
    ctx.fillStyle = '#9ca3af';
    ctx.font = '14px sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText('No data available', width / 2, height / 2);
    loading.value = false;
    return;
  }

  // Draw chart
  ctx.clearRect(0, 0, width, height);

  if (props.type === 'bar') {
    drawBarChart(ctx, width, height);
  } else {
    drawLineChart(ctx, width, height);
  }

  loading.value = false;
};

const drawLineChart = (ctx, width, height) => {
  const padding = 40;
  const chartWidth = width - padding * 2;
  const chartHeight = height - padding * 2;

  const values = props.data.map(d => d[props.dataKey]);
  const labels = props.data.map(d => d[props.labelKey]);

  const maxValue = Math.max(...values, 1);
  const minValue = Math.min(...values, 0);

  // Draw grid
  ctx.strokeStyle = 'rgba(156, 163, 175, 0.2)';
  ctx.lineWidth = 1;
  for (let i = 0; i <= 5; i++) {
    const y = padding + (chartHeight / 5) * (5 - i);
    ctx.beginPath();
    ctx.moveTo(padding, y);
    ctx.lineTo(width - padding, y);
    ctx.stroke();

    // Y-axis labels
    const value = minValue + ((maxValue - minValue) / 5) * i;
    ctx.fillStyle = '#6b7280';
    ctx.font = '10px sans-serif';
    ctx.textAlign = 'right';
    ctx.fillText(value.toFixed(1), padding - 5, y + 4);
  }

  // Draw line
  ctx.strokeStyle = props.color;
  ctx.lineWidth = 2;
  ctx.beginPath();

  values.forEach((value, index) => {
    const x = padding + (chartWidth / (values.length - 1)) * index;
    const y = padding + chartHeight - ((value - minValue) / (maxValue - minValue)) * chartHeight;

    if (index === 0) {
      ctx.moveTo(x, y);
    } else {
      ctx.lineTo(x, y);
    }

    // X-axis labels
    ctx.fillStyle = '#6b7280';
    ctx.font = '10px sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText(labels[index], x, height - padding + 20);
  });

  ctx.stroke();

  // Draw points
  values.forEach((value, index) => {
    const x = padding + (chartWidth / (values.length - 1)) * index;
    const y = padding + chartHeight - ((value - minValue) / (maxValue - minValue)) * chartHeight;

    ctx.fillStyle = props.color;
    ctx.beginPath();
    ctx.arc(x, y, 4, 0, Math.PI * 2);
    ctx.fill();
  });
};

const drawBarChart = (ctx, width, height) => {
  const padding = 40;
  const chartWidth = width - padding * 2;
  const chartHeight = height - padding * 2;

  const values = props.data.map(d => d[props.dataKey]);
  const labels = props.data.map(d => d[props.labelKey]);

  const maxValue = Math.max(...values, 1);
  const barWidth = (chartWidth / values.length) * 0.7;
  const gap = (chartWidth / values.length) * 0.3;

  // Draw grid
  ctx.strokeStyle = 'rgba(156, 163, 175, 0.2)';
  ctx.lineWidth = 1;
  for (let i = 0; i <= 5; i++) {
    const y = padding + (chartHeight / 5) * (5 - i);
    ctx.beginPath();
    ctx.moveTo(padding, y);
    ctx.lineTo(width - padding, y);
    ctx.stroke();

    // Y-axis labels
    const value = (maxValue / 5) * i;
    ctx.fillStyle = '#6b7280';
    ctx.font = '10px sans-serif';
    ctx.textAlign = 'right';
    ctx.fillText(value.toFixed(0), padding - 5, y + 4);
  }

  // Draw bars
  values.forEach((value, index) => {
    const x = padding + (barWidth + gap) * index + gap / 2;
    const barHeight = (value / maxValue) * chartHeight;
    const y = padding + chartHeight - barHeight;

    ctx.fillStyle = props.color;
    ctx.fillRect(x, y, barWidth, barHeight);

    // X-axis labels
    ctx.fillStyle = '#6b7280';
    ctx.font = '10px sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText(labels[index], x + barWidth / 2, height - padding + 20);
  });
};

watch(() => props.data, () => {
  initChart();
}, { deep: true });

onMounted(() => {
  initChart();
  window.addEventListener('resize', initChart);
});

onUnmounted(() => {
  window.removeEventListener('resize', initChart);
  if (chart) {
    chart.destroy();
  }
});
</script>
