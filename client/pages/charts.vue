<template>
    <div class="min-h-screen py-6 flex flex-col justify-center sm:py-12 flex-1">
      <div class="relative py-3 w-full px-3">
        <h1 class="text-3xl font-bold mb-5 text-center text-white">Cinema Ticket Statistics</h1>
        <MultiSelect 
          v-model="values" 
          name="multiselect" 
          :items="items"
          show-by="name"
          :init="init"
          chips-text-style="break-words whitespace-pre-wrap w-full flex flex-wrap overflow-hidden text-sm text-black-950"
          multiple
        />
        <div class="relative py-6 shadow-lg sm:rounded-3xl sm:p-5 w-full">
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
  
  const items = [
    {id: 1, name: 'Classic Margherita Pizza'},
    {id: 2, name: 'Vegetarian Stir-Fry'},
    {id: 3, name: 'Chocolate Chip Cookies'},
    {id: 4, name: 'Chicken Alfredo Pasta'},
    {id: 5, name: 'Mango Salsa Chicken'},
    {id: 6, name: 'Quinoa Salad with Avocado'},
    {id: 6, name: 'Tomato Basil Bruschetta'},
  ]

  const values = ref<any[]>([])
  // const {data, status, error} = useFetch<{recipes: any[]}>('https://dummyjson.com/recipes?limit=2&select=name', {
  //   pick: ['recipes'],
  //   cache: 'no-cache'
  // });
  // const init = computed(() => {
  //   console.log("Data is: ", data.value?.recipes)
  //   return data.value?.recipes
  // })
  const init = ref()
  console.log("BEFORE ONE SECOND", init.value)
  setTimeout(() => {
    init.value = [{id: 1, name: 'Classic Margherita Pizza'}]
    console.log("AFTER ONE SECOND", init.value)
  }, 1000);
  const { result } = useQuery(CHART_QUERY, {}, {
    fetchPolicy: 'no-cache'
  });
  watch(values, (value) => {
    console.log("Result is: ", value)
  })

  const dailyData = computed(() => {
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
  
  const categoryData = computed(() => {
    return genra.value?.group_movie_by_category.map((item:any) => ({
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

    labels: categoryData.value?.map((item:any) => item.category),
  }))

  const pieChartSeries = computed(() => 
    categoryData.value?.map((item:any) => item.count) || []
  )
</script>
<style>
.apexcharts-tooltip span {
    color: black;
}
</style>