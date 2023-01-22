# Инкремент 15

Добавьте в свой проект бенчмарки, измеряющие скорость выполнения важнейших компонентов вашей системы.
Проведите анализ использования памяти вашим проектом, определите и исправьте неэффективные части кода.


1. Запускаем бенчмарки: `make bench p=before`
2. Смотрим результат: `make pprof p=before`
3. Оптимизируем код
4. Запускаем бенчмарки снова:`make bench p=after`
5. Смотрим результат: `make pprof p=after`
6. Сравниваем результаты:`make pprof_diff p1=before p2=after`

## Результаты оптимизации
Использование указателей в структурах данных позволило сократить потребление памяти в 2 раза.
```
Type: alloc_space
Time: Jan 9, 2023 at 10:18pm (MSK)
Showing nodes accounting for -2725.60MB, 57.42% of 4746.82MB total
Dropped 48 nodes (cum <= 23.73MB)
      flat  flat%   sum%        cum   cum%
-1777.08MB 37.44% 37.44% -1777.08MB 37.44%  github.com/ofstudio/go-shortener/internal/repo.(*MemoryRepo).ShortURLGetByID
 -891.51MB 18.78% 56.22%  -891.51MB 18.78%  github.com/ofstudio/go-shortener/internal/repo.(*MemoryRepo).UserGetByID
  -57.01MB  1.20% 57.42%   -57.01MB  1.20%  github.com/ofstudio/go-shortener/internal/repo.(*MemoryRepo).ShortURLGetByUserID
         0     0% 57.42% -1776.58MB 37.43%  github.com/ofstudio/go-shortener/internal/repo.BenchmarkMemoryRepo_ShortURLGetByID
         0     0% 57.42%   -57.01MB  1.20%  github.com/ofstudio/go-shortener/internal/repo.BenchmarkMemoryRepo_ShortURLGetByUserID
         0     0% 57.42%  -892.01MB 18.79%  github.com/ofstudio/go-shortener/internal/repo.BenchmarkMemoryRepo_UserGetByID
         0     0% 57.42% -2725.60MB 57.42%  testing.(*B).launch
         0     0% 57.42% -2726.10MB 57.43%  testing.(*B).runN
```