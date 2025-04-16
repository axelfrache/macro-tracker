import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';

interface MacroChartProps {
  proteins: number;
  carbs: number;
  fats: number;
  fiber?: number;
}

export const MacroChart = ({ proteins, carbs, fats, fiber = 0 }: MacroChartProps) => {
  const data = [
    { name: 'Protéines', value: proteins },
    { name: 'Glucides', value: carbs },
    { name: 'Lipides', value: fats },
  ];

  if (fiber > 0) {
    data.push({ name: 'Fibres', value: fiber });
  }

  const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#9C27B0'];

  return (
    <ResponsiveContainer width="100%" height={300}>
      <PieChart>
        <Pie
          data={data}
          cx="50%"
          cy="50%"
          labelLine={false}
          outerRadius={80}
          fill="#8884d8"
          dataKey="value"
        >
          {data.map((_, index) => (
            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
          ))}
        </Pie>
        <Tooltip />
        <Legend />
      </PieChart>
    </ResponsiveContainer>
  );
};
