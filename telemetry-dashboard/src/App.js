import './App.css';
import RealTimeData from "./components/RealTimeData";
import HistoricalGraph from "./components/HistoricalGraph";

function App() {
  return (
      <div className="App">
          <RealTimeData />
          <HistoricalGraph />
      </div>
  );
}

export default App;
