import React from "react";

export default function Child2({ handleClick }) {
  return (
    <div>
      Child2
      <button onClick={() => handleClick(2)}>Child1 +2</button>
    </div>
  );
}
