let arr = [
    {name:'a', val: 1},
    {name:'b', val: 2},
    {name:'c', val: 3},
]

let arr1 = arr.filter(it => {
    return it.name !== 'c'
}).map(it => {
    it.val *= 2
    return it
})

console.log(arr1)
