<template>
  <div>
    <h3 class="font-weight-thin text-h3 mt-10">{{ total }} Orders</h3>
    <v-data-table
        :headers="tableHeaders"
        :items="tableItems"
        :loading="loading"
        :page="query.page"
        :items-per-page="query.limit_num"
        sort-by="calories"
        class="elevation-2 mt-15"
        @update:items-per-page="handleCount"
    >
      <template v-slot:item.CreatedAt="{ item }">
        <span>{{ new Date(item.CreatedAt * 1000).toLocaleString() }}</span>
      </template>
      <template v-slot:item.UpdatedAt="{ item }">
        <span>{{ new Date(item.UpdatedAt * 1000).toLocaleString() }}</span>
      </template>
      <template v-slot:top>
        <v-toolbar flat>
          <v-text-field
              v-model="query.name"
              append-icon="mdi-magnify"
              label="Search by name"
              class="pt-10"
              @keydown.enter="handleSearch"
          ></v-text-field>
          <v-spacer></v-spacer>
          <form-dialog @on-success="handleSearch"/>
        </v-toolbar>
      </template>

      <template v-slot:item.status="{item}">
        <v-chip
            color="red"
            v-if=" item.status == 0 "
        >
          Wait for payment
        </v-chip>
        <v-chip
            color="primary"
            v-else-if=" item.status == 1 "
        >
          Paid, please wait
        </v-chip>
        <v-chip
            color="green"
            v-else-if=" item.status == 2 "
        >
          Success
        </v-chip>
      </template>
      <template v-slot:item.action="{ item }">
        <v-btn x-medium rounded @click="pay(item)" :loading="paying[item.uuid]" v-if=" item.status == 0">
          <v-icon class="mr-3">mdi-wallet</v-icon>
          Pay
        </v-btn>
      </template>
      <template v-slot:no-data>No data</template>
    </v-data-table>
  </div>
</template>
<script>
import FormDialog from './components/form-dialog'
import { deleteZeroAccessClient, fetchZeroAccessClients, postZeroAccessClientsPayNotify } from '@/api'
import { payOrder, OrderPaid } from '@/utils/ethers'

export default {
  components: { FormDialog },
  data: () => ({
    loading: false,
    paying: [],
    query: {
      name: '',
      page: 1,
      limit_num: 15
    },
    tableHeaders: [
      { text: 'Name', align: 'start', sortable: true, value: 'name' },
      { text: 'OrderId', sortable: true, value: 'uuid' },
      { text: 'Listen port', sortable: true, value: 'port' },
      { text: 'Duration(Hours)', sortable: false, value: 'duration' },
      { text: 'Price(CSD)', sortable: true, value: 'price' },
      { text: 'Status', sortable: true, value: 'status' },
      { text: 'Action', value: 'action' },
      { text: 'Server', sortable: true, value: 'server_cid' },
      { text: 'PeerId', sortable: true, value: 'peer_id' },
      { text: 'Resource', sortable: true, value: 'resource_cid' },
      { text: 'Created at', sortable: true, value: 'CreatedAt' },
      { text: 'Updated at', sortable: true, value: 'UpdatedAt' }
    ],
    tableItems: [],
    total: 0
  }),
  created() {
    this.getTableItems()
  },
  methods: {
    handleSearch() {
      this.query.page = 1
      this.getTableItems()
    },
    async pay(item) {
      this.paying[item.uuid] = true
      const payStatus = await payOrder(item.name, item.duration, item.uuid, item.price, item.peer_id)
      if (payStatus === 'ok' || payStatus === OrderPaid) {
        postZeroAccessClientsPayNotify({ uuid: item.uuid }).then(res => {
          this.$emit('on-success')
          this.$message.success()
          this.paying[item.uuid] = false
          this.getTableItems()
        }).finally(() => {
          this.paying[item.uuid] = false
          this.loopResult()
        })
      } else {
        this.$message.error(payStatus)
        this.paying[item.uuid] = false
      }
    },
    getTableItems() {
      this.loading = true
      fetchZeroAccessClients(this.query).then(res => {
        this.tableItems = res.data.list || []
        this.total = res.data.paginate.total
        for (let i = 0; i < this.tableItems.length; i++) {
          this.paying[this.tableItems[i].uuid] = false
        }
      }).finally(() => {
        this.loading = false
      })
    },
    handleCount(v) {
      this.query.limit_num = v
      this.handleSearch()
    },
    handleDelete(ref) {
      const item = ref.data

      deleteZeroAccessClient(item.ID).then(_ => {
        ref.$close()
      }).finally(() => {
        this.handleSearch()
      })
    },
    loopResult() {
      let i = 0
      const timer = setInterval(() => {
        this.Interval(timer, i++)
      }, 10000)
    },
    Interval(timer, i) {
      setTimeout(() => {
        this.getTableItems()
        if (i >= 4) {
          clearInterval(timer)
        }
      }, 0)
    }
  }
}
</script>
