using Kafka.Example.Filters;
using Microsoft.AspNetCore.Mvc;

namespace Kafka.Example.Controllers
{
    [ApiController]
    [Route("api/[Controller]")]
    [ServiceFilter(typeof(TimeTracker))]
    [Delayer(3000)]
    public class ProductsController : ControllerBase
    {
        [HttpGet]
        public IActionResult Get() => NoContent();
        [HttpPost]
        public IActionResult Post() => NoContent();
        [HttpPut]
        public IActionResult Put() => NoContent();
        [HttpDelete]
        public IActionResult Delete() => NoContent();
    }
}