<template>
    <div class="min-h-screen py-6 flex flex-col justify-center sm:py-12 flex-1">
      <div class="relative py-3 w-full mx-3">
        <div class="relative py-6 shadow-lg sm:rounded-3xl sm:p-5 w-full">
          <h1 class="text-3xl font-bold mb-8 text-center text-blue-500">Cinema Ticket Statistics</h1>
          <ClientOnly fallback-tag="span" fallback="Loading ...">
              <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
                <!-- Daily Tickets Chart -->
                <div class="bg-gray-800 p-3 rounded-xl shadow-md text-white min-w-80">
                  <h2 class="text-xl font-semibold mb-4">Daily Ticket Sales</h2>
                      <apexchart
                        type="line"
                        height="350"
                        :options="dailyChartOptions"
                        :series="dailyChartSeries"
                      ></apexchart>
                </div>
                <div class="bg-gray-800 p-3 rounded-xl shadow-md text-white min-w-80">
                  <h2 class="text-xl font-semibold mb-4">Daily Ticket Sales</h2>
                      <apexchart
                        type="line"
                        height="350"
                        :options="dailyChartOptions"
                        :series="dailyChartSeries"
                      ></apexchart>
                </div>
                <!-- Movie Categories Pie Chart -->
                <div class="bg-gray-800 p-3 rounded-xl shadow-md text-white min-w-80">
                  <h2 class="text-xl font-semibold mb-4">Movie Categories</h2>
                      <apexchart
                        type="pie"
                        height="350"
                        :options="pieChartOptions"
                        :series="pieChartSeries"
                      ></apexchart>
                </div>
              </div>
          </ClientOnly>
        </div>
      </div>
    </div>
  </template>
  
  <script setup lang="ts">
  
  import { CHART_GENRE_QUERY, CHART_QUERY } from '../graphql/queries/chart';
  // Sample data for daily ticket sales
  const { result } = useQuery(CHART_QUERY, {}, {
    fetchPolicy: 'no-cache'
  });
  
  const dailyData = computed(() => {
    console.log(result.value)
    return result.value?.tickets_graph.map((item:any) => ({
      date: item.t_date,
      price: item.total_price,
      count: item.count
    })) || []
  })
  

  const { result:genra } = useQuery(CHART_GENRE_QUERY, {}, {
    fetchPolicy: 'no-cache'
  })
  
  // Sample data for movie categories
  
  // const categoryData = ref([
  //   { category: 'Action', count: 50 },
  //   { category: 'Comedy', count: 30 },
  //   { category: 'Drama', count: 20 },
  //   { category: 'Sci-Fi', count: 25 },
  //   { category: 'Horror', count: 15 },
  // ])
  const categoryData = computed(() => {
    return genra.value.group_movie_by_category.map((item:any) => ({
      category: item.genre,
      count: item.count
    }))
  })
  // Daily chart options and series
  const dailyChartOptions = computed(() => ({
    chart: {
      id: 'daily-sales',
      foreColor: '#fff',
      toolbar: {
        show: false
      }
    },
    xaxis: {
      categories: dailyData.value.map((item:any) => item.date),
      labels: {
        rotate: -45,
        rotateAlways: true
      }
    },
    yaxis: [
      {
        title: {
          text: 'Tickets Sold'
        }
      },
      {
        opposite: true,
        title: {
          text: 'Revenue ($)'
        }
      }
    ],
    colors: ['#3B82F6', '#10B981'],
    stroke: {
      curve: 'smooth'
    },
    legend: {
      position: 'top'
    }
  }))
  
  const dailyChartSeries = computed(() => [
    {
      name: 'Tickets Sold',
      type: 'line',
      data: dailyData.value.map((item:any) => item.count)
    },
    {
      name: 'Revenue',
      type: 'line',
      data: dailyData.value.map((item:any) => item.price)
    }
  ])
  
  // Pie chart options and series
  const pieChartOptions = computed(() => ({
    chart: {
      id: 'category-distribution',
      foreColor: "#fff",
    },

    labels: categoryData.value.map((item:any) => item.category),
  }))

  const pieChartSeries = computed(() => 
    categoryData.value.map((item:any) => item.count)
  )
</script>
<style>
.apexcharts-tooltip span {
    color: black;
}
</style>