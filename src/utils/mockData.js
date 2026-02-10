// 模拟考试详情数据
export const mockExamDetail = {
  questions: [
    {
      id: 1,
      question: '以下哪个不是JavaScript的数据类型？',
      type: '单选题',
      studentAnswer: 'd',
      correctAnswer: 'd',
      score: 10,
      correct: true,
      difficulty: '简单',
      knowledgePoint: '基础语法'
    },
    {
      id: 2,
      question: '请解释什么是闭包(Closure)？',
      type: '简答题',
      studentAnswer: '闭包是指有权访问另一个函数作用域中变量的函数',
      correctAnswer: '闭包是指有权访问另一个函数作用域中变量的函数',
      score: 18,
      correct: true,
      difficulty: '中等',
      knowledgePoint: '函数进阶'
    },
    {
      id: 3,
      question: '以下代码的输出结果是什么？',
      type: '编程题',
      studentAnswer: '输出: 1, 2, 3',
      correctAnswer: '输出: 1, 2, 3',
      score: 25,
      correct: true,
      difficulty: '困难',
      knowledgePoint: '异步编程'
    }
  ]
}

// 模拟WebSocket重连配置
export const websocketConfig = {
  maxRetries: 5,
  retryInterval: 3000,
  timeout: 5000
}

// 模拟图表配置
export const chartConfig = {
  colors: {
    primary: '#409eff',
    success: '#67c23a',
    warning: '#e6a23c',
    danger: '#f56c6c'
  },
  animationDuration: 300,
  responsive: true
}
