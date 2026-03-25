// OODA loop row templates for go-risk-it dashboards.
// Each dashboard follows the Observe-Orient-Decide-Act structure.
{
  observeRow():: {
    title: 'Observe \u2014 Am I OK?',
    type: 'row',
    collapsed: false,
    id: 100,
  },

  orientRow():: {
    title: 'Orient \u2014 What\u2019s the shape?',
    type: 'row',
    collapsed: false,
    id: 200,
  },

  decideRow():: {
    title: 'Decide \u2014 Where\u2019s the bottleneck?',
    type: 'row',
    collapsed: false,
    id: 300,
  },

  actRow():: {
    title: 'Act \u2014 What\u2019s the evidence?',
    type: 'row',
    collapsed: false,
    id: 400,
  },
}
